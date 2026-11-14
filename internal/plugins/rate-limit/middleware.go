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

type middleware struct {
	RateLimiter RateLimiter
	lastLogged  atomic.Int64 // unix nanoseconds of the last store error logged
}

// NewMiddleware returns the rate-limit plugin: it counts requests per client
// IP with the given limiter and answers 429 once the limit is exceeded.
//
// If the limiter's store fails, for example because Redis is down, requests
// are let through without a limit rather than rejected, so an outage of the
// store does not take down every service that uses the plugin.
func NewMiddleware(rl RateLimiter) func(http.Handler) http.Handler {
	return (&middleware{RateLimiter: rl}).wrap
}

func (rlm *middleware) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, remaining, resetIn, err := rlm.RateLimiter.RateLimit(rateLimitKeyPrefix + clientIP(req))
		switch {
		case errors.Is(err, ErrRateLimitExceeded):
			reset := seconds(resetIn)
			rlm.setHeaders(w, 0, reset)
			w.Header().Set("Retry-After", reset)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		case err != nil:
			rlm.logStoreError(err)
		default:
			rlm.setHeaders(w, remaining, seconds(resetIn))
		}
		next.ServeHTTP(w, req)
	})
}

// setHeaders reports the client's limit, what is left of it, and the seconds
// until it resets.
func (rlm *middleware) setHeaders(w http.ResponseWriter, remaining int64, reset string) {
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rlm.RateLimiter.GetLimit(), 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	w.Header().Set("X-RateLimit-Reset", reset)
}

func (rlm *middleware) logStoreError(err error) {
	now := time.Now().UnixNano()
	last := rlm.lastLogged.Load()
	if now-last < int64(storeErrorLogInterval) || !rlm.lastLogged.CompareAndSwap(last, now) {
		return
	}
	log.Printf("rate-limit: store failed, letting requests through without a limit: %v", err)
}

// seconds rounds d up to whole seconds, so a client that waits that long is
// past the reset.
func seconds(d time.Duration) string {
	return strconv.Itoa(int(math.Ceil(d.Seconds())))
}
