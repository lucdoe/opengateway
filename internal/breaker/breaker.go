// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package breaker stops the gateway from sending requests to an upstream
// that keeps failing, so a failing service answers fast instead of tying up
// connections until its timeout.
package breaker

import (
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/lucdoe/opengateway/internal/config"
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
	retryAfter := strconv.Itoa(int(math.Ceil(settings.OpenTimeout.Seconds())))

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
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
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
