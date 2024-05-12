// rate-limit.go in middleware package
package middleware

import (
	"net/http"
	"strconv"

	ratelimit "github.com/lucdoe/open-gateway/gateway/internal/plugins/rate-limit"
)

type RateLimitMiddleware struct {
	RateLimiter ratelimit.RateLimiter
}

func NewRateLimitMiddleware(rl ratelimit.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		RateLimiter: rl,
	}
}

func (rlm *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key := req.RemoteAddr
		_, remaining, window, err := rlm.RateLimiter.RateLimit(key)
		if err != nil {
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rlm.RateLimiter.(*ratelimit.RateLimitService).Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			w.Header().Set("X-RateLimit-Reset", window.String())

			if err.Error() == "rate limit exceeded" {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rlm.RateLimiter.(*ratelimit.RateLimitService).Limit, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		w.Header().Set("X-RateLimit-Reset", window.String())

		next.ServeHTTP(w, req)
	})
}
