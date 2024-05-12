package ratelimit

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

type Cache interface {
	Increment(key string, window time.Duration) (int64, error)
}

type RateLimiter interface {
	Middleware(next http.Handler) http.Handler
}

type RateLimitService struct {
	Store  Cache
	Limit  int64
	Window time.Duration
}

func NewRateLimitService(store Cache, limit int64, window time.Duration) *RateLimitService {
	return &RateLimitService{
		Store:  store,
		Limit:  limit,
		Window: window,
	}
}

func (r *RateLimitService) RateLimit(key string) (count int64, remaining int64, window time.Duration, err error) {
	curCount, err := r.Store.Increment(key, r.Window)
	if err != nil {
		return 0, 0, 0, err
	}

	curRemaining := r.Limit - curCount
	if curRemaining < 0 {
		return curCount, 0, 0, errors.New("rate limit exceeded")
	}

	return curCount, curRemaining, r.Window, nil
}

func (r *RateLimitService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key := req.RemoteAddr
		_, remaining, window, err := r.RateLimit(key)
		if err != nil {
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(r.Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			w.Header().Set("X-RateLimit-Reset", window.String())

			if err.Error() == "redis error" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Error(w, err.Error(), http.StatusTooManyRequests)

			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(r.Limit, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		w.Header().Set("X-RateLimit-Reset", window.String())

		next.ServeHTTP(w, req)
	})
}
