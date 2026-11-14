// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveLogged sends one request through the logger plugin on a router with
// the given route and returns the line it logged.
func serveLogged(t *testing.T, routePath, target string, handler http.HandlerFunc) map[string]any {
	t.Helper()
	var out syncBuffer
	l := newLogger(&out, nil)

	router := mux.NewRouter()
	router.Handle(routePath, NewMiddleware(l)(handler))
	req := httptest.NewRequest("GET", target, nil)
	req.RemoteAddr = "203.0.113.5:51000"
	router.ServeHTTP(httptest.NewRecorder(), req)
	require.NoError(t, l.Close())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.Len(t, lines, 1)
	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &entry))
	return entry
}

func TestMiddlewareLogsTheAnsweredRequest(t *testing.T) {
	entry := serveLogged(t, "/api/people/{id}", "http://gateway.test/api/people/7?token=secret",
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Millisecond)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Bad gateway\n"))
		})

	assert.Equal(t, "request", entry["msg"])
	assert.Equal(t, "GET", entry["method"])
	assert.Equal(t, "/api/people/7", entry["path"], "the query string is left out")
	assert.Equal(t, "/api/people/{id}", entry["route"])
	assert.Equal(t, float64(http.StatusBadGateway), entry["status"])
	assert.Equal(t, float64(len("Bad gateway\n")), entry["bytes"])
	assert.Equal(t, "203.0.113.5", entry["client"])
	assert.GreaterOrEqual(t, entry["duration_ms"], 5.0)
	assert.NotContains(t, entry, "token")
}

func TestMiddlewareLogsOKForAHandlerThatWritesNothing(t *testing.T) {
	entry := serveLogged(t, "/ping", "http://gateway.test/ping", func(w http.ResponseWriter, r *http.Request) {})

	assert.Equal(t, float64(http.StatusOK), entry["status"])
	assert.Equal(t, float64(0), entry["bytes"])
}

func TestMiddlewareLogsWithoutARouter(t *testing.T) {
	var out syncBuffer
	l := newLogger(&out, nil)
	req := httptest.NewRequest("GET", "http://gateway.test/x", nil)
	req.RemoteAddr = "not-an-address"

	NewMiddleware(l)(http.NotFoundHandler()).ServeHTTP(httptest.NewRecorder(), req)
	require.NoError(t, l.Close())

	assert.Contains(t, out.String(), `"route":""`)
	assert.Contains(t, out.String(), `"client":"not-an-address"`)
	assert.Contains(t, out.String(), `"status":404`)
}

func TestMillisecondsIsPreciseToTheMicrosecond(t *testing.T) {
	assert.Equal(t, 12.345, milliseconds(12345678*time.Nanosecond))
}
