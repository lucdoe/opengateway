// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package metrics counts what the gateway answers and serves it in the
// Prometheus text format, without a client library.
package metrics

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lucdoe/opengateway/internal/statuswriter"
)

// buckets are the upper bounds, in seconds, of the duration histogram.
var buckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// Breaker states, as the circuit breaker gauge reports them.
const (
	BreakerClosed   = 0
	BreakerHalfOpen = 1
	BreakerOpen     = 2
)

// Registry holds every metric. It is safe for concurrent use.
type Registry struct {
	mu       sync.Mutex
	routes   []*route
	breakers map[string]*atomic.Int64
}

// route holds the metrics of one route. Counting a request takes no lock
// once its method and status have been seen.
type route struct {
	service, path string
	requests      sync.Map // "METHOD status" → *atomic.Uint64
	buckets       []atomic.Uint64
	count         atomic.Uint64
	sumMicros     atomic.Uint64
}

func New() *Registry {
	return &Registry{breakers: map[string]*atomic.Int64{}}
}

// Middleware counts every request the route answers, whoever answers it,
// labelled with the service and the route's path as configured, not the
// request path, so the number of series stays bounded.
func (m *Registry) Middleware(service, path string) func(http.Handler) http.Handler {
	r := &route{service: service, path: path, buckets: make([]atomic.Uint64, len(buckets))}
	m.mu.Lock()
	m.routes = append(m.routes, r)
	m.mu.Unlock()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			rec := statuswriter.New(w)
			next.ServeHTTP(rec, req)
			r.observe(req.Method, rec.Status(), time.Since(start))
		})
	}
}

func (r *route) observe(method string, status int, took time.Duration) {
	if status == 0 {
		status = http.StatusOK // the handler wrote nothing
	}
	key := method + " " + strconv.Itoa(status)
	counter, ok := r.requests.Load(key)
	if !ok {
		counter, _ = r.requests.LoadOrStore(key, new(atomic.Uint64))
	}
	counter.(*atomic.Uint64).Add(1)

	seconds := took.Seconds()
	for i, bound := range buckets {
		if seconds <= bound {
			r.buckets[i].Add(1)
		}
	}
	r.count.Add(1)
	r.sumMicros.Add(uint64(took.Microseconds()))
}

// BreakerState records the state of a service's circuit breaker:
// BreakerClosed, BreakerHalfOpen or BreakerOpen.
func (m *Registry) BreakerState(service string, state int64) {
	m.mu.Lock()
	gauge, ok := m.breakers[service]
	if !ok {
		gauge = new(atomic.Int64)
		m.breakers[service] = gauge
	}
	m.mu.Unlock()
	gauge.Store(state)
}

// ServeHTTP writes every metric in the Prometheus text format, sorted so
// the output is stable.
func (m *Registry) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	m.WriteTo(w)
}

// WriteTo writes every metric in the Prometheus text format.
func (m *Registry) WriteTo(w io.Writer) (int64, error) {
	m.mu.Lock()
	routes := append([]*route(nil), m.routes...)
	breakers := make(map[string]int64, len(m.breakers))
	for service, gauge := range m.breakers {
		breakers[service] = gauge.Load()
	}
	m.mu.Unlock()
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].service+" "+routes[i].path < routes[j].service+" "+routes[j].path
	})

	var b strings.Builder
	b.WriteString("# HELP opengateway_requests_total Requests answered, by service, route, method and status.\n")
	b.WriteString("# TYPE opengateway_requests_total counter\n")
	for _, r := range routes {
		var keys []string
		r.requests.Range(func(key, _ any) bool { keys = append(keys, key.(string)); return true })
		sort.Strings(keys)
		for _, key := range keys {
			method, status, _ := strings.Cut(key, " ")
			counter, _ := r.requests.Load(key)
			fmt.Fprintf(&b, "opengateway_requests_total{%s,method=%s,status=%s} %d\n",
				r.labels(), quote(method), quote(status), counter.(*atomic.Uint64).Load())
		}
	}

	b.WriteString("# HELP opengateway_request_duration_seconds Time to answer a request, plugins included.\n")
	b.WriteString("# TYPE opengateway_request_duration_seconds histogram\n")
	for _, r := range routes {
		for i, bound := range buckets {
			fmt.Fprintf(&b, "opengateway_request_duration_seconds_bucket{%s,le=%s} %d\n",
				r.labels(), quote(strconv.FormatFloat(bound, 'g', -1, 64)), r.buckets[i].Load())
		}
		count := r.count.Load()
		fmt.Fprintf(&b, "opengateway_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", r.labels(), count)
		fmt.Fprintf(&b, "opengateway_request_duration_seconds_sum{%s} %s\n", r.labels(), formatFloat(float64(r.sumMicros.Load())/1e6))
		fmt.Fprintf(&b, "opengateway_request_duration_seconds_count{%s} %d\n", r.labels(), count)
	}

	b.WriteString("# HELP opengateway_circuit_breaker_state State of a service's circuit breaker: 0 closed, 1 half-open, 2 open.\n")
	b.WriteString("# TYPE opengateway_circuit_breaker_state gauge\n")
	services := make([]string, 0, len(breakers))
	for service := range breakers {
		services = append(services, service)
	}
	sort.Strings(services)
	for _, service := range services {
		fmt.Fprintf(&b, "opengateway_circuit_breaker_state{service=%s} %d\n", quote(service), breakers[service])
	}

	n, err := io.WriteString(w, b.String())
	return int64(n), err
}

func (r *route) labels() string {
	return "service=" + quote(r.service) + ",route=" + quote(r.path)
}

// quote writes a label value as the text format wants it.
func quote(value string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(value) + `"`
}

func formatFloat(f float64) string {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return "0"
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}
