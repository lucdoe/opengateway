// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package breaker stops the gateway from sending requests to an upstream
// that keeps failing, so a failing service answers fast instead of tying up
// connections until its timeout.
package breaker

import (
	"errors"
	"log"
	"net/http"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/httpx"
	"github.com/lucdoe/opengateway/internal/statuswriter"
	"github.com/sony/gobreaker/v2"
)

var (
	errUpstream   = errors.New("upstream failed")
	errClientGone = errors.New("client went away")
)

// Wrap returns a handler that opens the breaker after
// settings.ConsecutiveFailures failures in a row. A failure is a 500, 502, 503
// or 504; the proxy itself answers 502 and 504 for an unreachable or slow
// upstream.
//
// While open, requests get 503 with a Retry-After header for
// settings.OpenTimeout. Then up to settings.HalfOpenRequests trial requests
// decide whether the breaker closes or opens again. Requests the client
// abandons are not counted. name labels the log line for each state change.
func Wrap(name string, settings config.CircuitBreakerConfig, next http.Handler) http.Handler {
	threshold := uint32(settings.ConsecutiveFailures)
	cb := gobreaker.NewCircuitBreaker[struct{}](gobreaker.Settings{
		Name:        name,
		MaxRequests: uint32(settings.HalfOpenRequests),
		Timeout:     settings.OpenTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= threshold
		},
		IsExcluded: func(err error) bool {
			return errors.Is(err, errClientGone)
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("circuit breaker %q: %s -> %s", name, from, to)
		},
	})
	retryAfter := httpx.Seconds(settings.OpenTimeout)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := cb.Execute(func() (struct{}, error) {
			rec := statuswriter.New(w)
			next.ServeHTTP(rec, r)
			if r.Context().Err() != nil {
				return struct{}{}, errClientGone
			}
			if isFailure(rec.Status()) {
				return struct{}{}, errUpstream
			}
			return struct{}{}, nil
		})
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			w.Header().Set("Retry-After", retryAfter)
			httpx.Error(w, http.StatusServiceUnavailable, "Service unavailable")
		}
	})
}

func isFailure(status int) bool {
	switch status {
	case http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}
