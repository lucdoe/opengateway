package ratelimit

import (
	"errors"
	"time"
)

type RateLimiter interface {
	Increment(key string, window time.Duration) (int64, error)
	Expire(key string, window time.Duration) error
}

type RateLimitService struct {
	Store  RateLimiter
	Limit  int64
	Window time.Duration
}

func NewRateLimitService(store RateLimiter, limit int64, window time.Duration) *RateLimitService {
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
		return count, 0, 0, errors.New("rate limit exceeded")
	}

	return curCount, curRemaining, r.Window, nil
}
