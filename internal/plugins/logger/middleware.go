// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/statuswriter"
)

// NewMiddleware returns the logger plugin. Once a request has been answered
// it logs the method, path, matched route, status, duration in milliseconds,
// response body bytes and client IP. The query string is not logged, since
// it can carry tokens.
func NewMiddleware(l *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := statuswriter.New(w)
			next.ServeHTTP(rec, r)

			status := rec.Status()
			if status == 0 {
				status = http.StatusOK // the handler wrote nothing
			}
			l.Request(r.Context(),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("route", route(r)),
				slog.Int("status", status),
				slog.Float64("duration_ms", milliseconds(time.Since(start))),
				slog.Int64("bytes", rec.Bytes()),
				slog.String("client", clientIP(r)),
			)
		})
	}
}

// route is the path template the request matched, such as /api/people/{id},
// so lines can be grouped by endpoint.
func route(r *http.Request) string {
	if current := mux.CurrentRoute(r); current != nil {
		if template, err := current.GetPathTemplate(); err == nil {
			return template
		}
	}
	return ""
}

// milliseconds is d in milliseconds, to the microsecond.
func milliseconds(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
