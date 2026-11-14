// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

const rateLimitKeyPrefix = "ratelimit:"

// storeErrorLogInterval is the least time between two log lines about a
// failing store, so an outage does not write one line per request.
const storeErrorLogInterval = 10 * time.Second

// NewMiddleware returns the rate-limit plugin: it counts requests per client
// IP with the given limiter and answers 429 once the limit is exceeded.
//
// If the limiter's store fails, for example because Redis is down, requests
// are let through without a limit rather than rejected, so an outage of the
// store does not take down every service that uses the plugin.
func NewMiddleware(l *Limiter) func(http.Handler) http.Handler {
	limit := strconv.FormatInt(l.limit, 10)
	var lastLogged atomic.Int64 // unix nanoseconds of the last store error logged
	setHeaders := func(w http.ResponseWriter, remaining int64, reset string) {
		w.Header().Set("X-RateLimit-Limit", limit)
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		w.Header().Set("X-RateLimit-Reset", reset)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			remaining, resetIn, err := l.Take(rateLimitKeyPrefix + clientIP(req))
			switch {
			case errors.Is(err, errLimitExceeded):
				reset := seconds(resetIn)
				setHeaders(w, 0, reset)
				w.Header().Set("Retry-After", reset)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			case err != nil:
				now := time.Now().UnixNano()
				if last := lastLogged.Load(); now-last >= int64(storeErrorLogInterval) && lastLogged.CompareAndSwap(last, now) {
					log.Printf("rate-limit: store failed, letting requests through without a limit: %v", err)
				}
			default:
				setHeaders(w, remaining, seconds(resetIn))
			}
			next.ServeHTTP(w, req)
		})
	}
}

// seconds rounds d up to whole seconds, so a client that waits that long is
// past the reset.
func seconds(d time.Duration) string {
	return strconv.Itoa(int(math.Ceil(d.Seconds())))
}
