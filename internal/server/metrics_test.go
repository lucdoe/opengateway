// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scrape(t *testing.T, s *Server) string {
	t.Helper()
	rec := httptest.NewRecorder()
	s.Metrics.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	return rec.Body.String()
}

func TestMetricsAreOffWithoutAMetricsPort(t *testing.T) {
	s, err := NewServer(&config.Config{Services: map[string]config.Service{
		"api": {URL: "http://api", Endpoints: []config.Endpoint{{HTTPMethod: "GET", Path: "/items"}}},
	}}, mux.NewRouter(), okProxy, Middlewares{}, nil)
	require.NoError(t, err)

	assert.Nil(t, s.Metrics)
	assert.Empty(t, s.MetricsAddr)
}

func TestMetricsCountEveryRouteAndBreaker(t *testing.T) {
	failing := func(string, time.Duration) (http.Handler, error) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadGateway) }), nil
	}
	cfg := &config.Config{
		Server: config.ServerConfig{MetricsPort: 9090, CircuitBreaker: config.CircuitBreakerConfig{ConsecutiveFailures: 1, OpenTimeout: time.Minute, HalfOpenRequests: 1}},
		Services: map[string]config.Service{
			"api": {URL: "http://api", Subpath: "/api", Endpoints: []config.Endpoint{{HTTPMethod: "GET", Path: "/items/{id}"}}},
		},
	}
	s, err := NewServer(cfg, mux.NewRouter(), failing, Middlewares{}, nil)
	require.NoError(t, err)
	assert.Equal(t, ":9090", s.MetricsAddr)
	assert.Contains(t, scrape(t, s), `opengateway_circuit_breaker_state{service="api"} 0`, "every breaker is there from the start")

	serve(s, "GET", "/api/items/7") // 502, opens the breaker
	serve(s, "GET", "/api/items/8") // 503 from the open breaker

	out := scrape(t, s)
	assert.Contains(t, out, `opengateway_requests_total{service="api",route="/api/items/{id}",method="GET",status="502"} 1`)
	assert.Contains(t, out, `opengateway_requests_total{service="api",route="/api/items/{id}",method="GET",status="503"} 1`, "answers of the gateway itself count too")
	assert.Contains(t, out, `opengateway_circuit_breaker_state{service="api"} 2`)
}

func TestRunServesMetricsOnTheirOwnPort(t *testing.T) {
	s, err := NewServer(&config.Config{
		Server:   config.ServerConfig{MetricsPort: 9090, ShutdownTimeout: time.Second},
		Services: map[string]config.Service{"api": {URL: "http://api", Endpoints: []config.Endpoint{{HTTPMethod: "GET", Path: "/items"}}}},
	}, mux.NewRouter(), okProxy, Middlewares{}, nil)
	require.NoError(t, err)
	s.Addr, s.MetricsAddr = freeAddr(t), freeAddr(t)

	ctx, stop := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()

	var body string
	require.Eventually(t, func() bool {
		res, err := http.Get("http://" + s.MetricsAddr + "/metrics")
		if err != nil {
			return false
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		body = string(b)
		return res.StatusCode == http.StatusOK
	}, 2*time.Second, 10*time.Millisecond)
	assert.Contains(t, body, "# TYPE opengateway_requests_total counter")

	res, err := http.Get("http://" + s.Addr + "/metrics")
	require.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode, "the gateway port does not serve metrics")

	stop()
	assert.NoError(t, <-done)
	_, err = http.Get("http://" + s.MetricsAddr + "/metrics")
	assert.Error(t, err, "metrics stop with the gateway")
}

func TestRunFailsWhenTheMetricsPortIsTaken(t *testing.T) {
	taken, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { taken.Close() })
	s := &Server{Addr: freeAddr(t), MetricsAddr: taken.Addr().String(), Router: mux.NewRouter()}

	err = s.Run(context.Background())

	assert.ErrorContains(t, err, "metrics:")
	assert.True(t, strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind"))
}

// freeAddr returns a local address no one listens on right now.
func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	l.Close()
	return addr
}

func TestServeMetricsLogsWhenItsListenerFails(t *testing.T) {
	logs := &lockedBuffer{}
	log.SetOutput(logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	closed, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	closed.Close()

	stop := (&Server{Metrics: metrics.New()}).serveMetrics(closed)
	defer stop()

	assert.Eventually(t, func() bool { return strings.Contains(logs.String(), "metrics:") }, time.Second, 5*time.Millisecond)
}

// lockedBuffer is a log output that one goroutine can write while another
// reads it.
type lockedBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}
