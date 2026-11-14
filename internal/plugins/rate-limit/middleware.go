// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"errors"
	"log"
	"net/http"
	"net/netip"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/lucdoe/opengateway/internal/httpx"
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
			remaining, resetIn, err := l.Take(rateLimitKeyPrefix + clientKey(httpx.ClientIP(req)))
			switch {
			case errors.Is(err, errLimitExceeded):
				reset := httpx.Seconds(resetIn)
				setHeaders(w, 0, reset)
				w.Header().Set("Retry-After", reset)
				httpx.Error(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			case err != nil:
				now := time.Now().UnixNano()
				if last := lastLogged.Load(); now-last >= int64(storeErrorLogInterval) && lastLogged.CompareAndSwap(last, now) {
					log.Printf("rate-limit: store failed, letting requests through without a limit: %v", err)
				}
			default:
				setHeaders(w, remaining, httpx.Seconds(resetIn))
			}
			next.ServeHTTP(w, req)
		})
	}
}

// clientKey groups an IPv6 client by its /64, the block one subscriber
// usually gets, so rotating addresses inside it does not reset the limit.
func clientKey(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is6() || addr.Is4In6() {
		return ip
	}
	return netip.PrefixFrom(addr, 64).Masked().String()
}
