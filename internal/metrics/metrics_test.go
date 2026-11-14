// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func scrape(m *Registry) string {
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	return rec.Body.String()
}

func TestMiddlewareCountsByMethodAndStatus(t *testing.T) {
	m := New()
	answer := func(status int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) })
	}
	items := m.Middleware("api", "/items/{id}")
	silent := m.Middleware("api", "/ping")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	for _, tc := range []struct {
		method string
		status int
	}{{"GET", 200}, {"GET", 200}, {"GET", 404}, {"POST", 429}} {
		items(answer(tc.status)).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(tc.method, "/items/7", nil))
	}
	silent.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ping", nil))

	out := scrape(m)
	assert.Contains(t, out, "# TYPE opengateway_requests_total counter\n")
	assert.Contains(t, out, `opengateway_requests_total{service="api",route="/items/{id}",method="GET",status="200"} 2`+"\n")
	assert.Contains(t, out, `opengateway_requests_total{service="api",route="/items/{id}",method="GET",status="404"} 1`+"\n")
	assert.Contains(t, out, `opengateway_requests_total{service="api",route="/items/{id}",method="POST",status="429"} 1`+"\n")
	assert.Contains(t, out, `opengateway_requests_total{service="api",route="/ping",method="GET",status="200"} 1`+"\n", "a handler that writes nothing answered 200")
	assert.Less(t, strings.Index(out, `route="/items/{id}"`), strings.Index(out, `route="/ping"`), "sorted")
}

func TestTheHistogramIsCumulative(t *testing.T) {
	r := &route{service: "api", path: "/x", buckets: make([]atomic.Uint64, len(buckets))}
	r.observe("GET", 200, 3*time.Millisecond)   // 0.005 and up
	r.observe("GET", 200, 300*time.Millisecond) // 0.5 and up
	r.observe("GET", 200, 20*time.Second)       // only +Inf
	m := &Registry{routes: []*route{r}, breakers: map[string]*atomic.Int64{}}

	out := scrape(m)
	for _, want := range []string{
		`opengateway_request_duration_seconds_bucket{service="api",route="/x",le="0.001"} 0`,
		`opengateway_request_duration_seconds_bucket{service="api",route="/x",le="0.005"} 1`,
		`opengateway_request_duration_seconds_bucket{service="api",route="/x",le="0.25"} 1`,
		`opengateway_request_duration_seconds_bucket{service="api",route="/x",le="0.5"} 2`,
		`opengateway_request_duration_seconds_bucket{service="api",route="/x",le="10"} 2`,
		`opengateway_request_duration_seconds_bucket{service="api",route="/x",le="+Inf"} 3`,
		`opengateway_request_duration_seconds_sum{service="api",route="/x"} 20.303`,
		`opengateway_request_duration_seconds_count{service="api",route="/x"} 3`,
	} {
		assert.Contains(t, out, want+"\n")
	}
}

func TestBreakerStateIsAGauge(t *testing.T) {
	m := New()
	m.BreakerState("payments", BreakerClosed)
	m.BreakerState("accounts", BreakerClosed)
	m.BreakerState("payments", BreakerOpen)

	out := scrape(m)

	assert.Contains(t, out, "# TYPE opengateway_circuit_breaker_state gauge\n")
	assert.Contains(t, out, `opengateway_circuit_breaker_state{service="accounts"} 0`+"\n")
	assert.Contains(t, out, `opengateway_circuit_breaker_state{service="payments"} 2`+"\n")
}

func TestLabelValuesAreEscaped(t *testing.T) {
	assert.Equal(t, `"a\\b\"c\nd"`, quote("a\\b\"c\nd"))
}

func TestServeHTTPSetsTheTextFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	New().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))

	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", rec.Header().Get("Content-Type"))
}

func TestFormatFloatNeverWritesInfOrNaN(t *testing.T) {
	assert.Equal(t, "1.5", formatFloat(1.5))
	assert.Equal(t, "0", formatFloat(math.Inf(1)))
}
