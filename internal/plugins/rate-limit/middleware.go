// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
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

// KeyFunc returns what a request is counted by.
type KeyFunc func(*http.Request) string

// ByClientIP counts per client IP, and IPv6 clients per /64.
func ByClientIP(r *http.Request) string {
	return clientKey(httpx.ClientIP(r))
}

// ByValue counts per value(r), such as a header or a token claim, and per
// client IP for requests without one, so leaving it out does not escape the
// limit. Values are stored hashed: they can be API keys, and a long one must
// not make a long store key. kind keeps equal values of different kinds,
// and IPs, apart.
func ByValue(kind string, value func(*http.Request) string) KeyFunc {
	return func(r *http.Request) string {
		v := value(r)
		if v == "" {
			return ByClientIP(r)
		}
		sum := sha256.Sum256([]byte(v))
		return kind + ":" + hex.EncodeToString(sum[:16])
	}
}

// ByHeader counts per value of the named header. Only use a header a proxy
// in front of the gateway sets, such as CF-Connecting-IP behind Cloudflare:
// a client can send any value, and with it as many keys as it likes.
func ByHeader(name string) KeyFunc {
	return ByValue("header", func(r *http.Request) string { return r.Header.Get(name) })
}

// NewMiddleware returns the rate-limit plugin: it counts requests per key
// with the given limiter and answers 429 once the limit is exceeded.
//
// If the limiter's store fails, for example because Redis is down, requests
// are let through without a limit rather than rejected, so an outage of the
// store does not take down every service that uses the plugin.
func NewMiddleware(l *Limiter, key KeyFunc) func(http.Handler) http.Handler {
	limit := strconv.FormatInt(l.limit, 10)
	var lastLogged atomic.Int64 // unix nanoseconds of the last store error logged
	setHeaders := func(w http.ResponseWriter, remaining int64, reset string) {
		w.Header().Set("X-RateLimit-Limit", limit)
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		w.Header().Set("X-RateLimit-Reset", reset)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			remaining, resetIn, err := l.Take(rateLimitKeyPrefix + key(req))
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
