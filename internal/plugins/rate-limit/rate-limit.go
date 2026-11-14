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
	"time"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// Cache is the store the counters live in.
type Cache interface {
	// Increment adds one to the counter at key and returns the new count and
	// the time left until it resets. A counter lives for window from its
	// first increment.
	Increment(key string, window time.Duration) (count int64, resetIn time.Duration, err error)
}

type RateLimiter interface {
	// RateLimit counts one request for key. remaining is how many more are
	// allowed and resetIn how long until the count starts over. It returns
	// ErrRateLimitExceeded once the limit is used up, and the store's error
	// if counting failed.
	RateLimit(key string) (count int64, remaining int64, resetIn time.Duration, err error)
	GetLimit() int64
}

type RateLimitConfig struct {
	Store  Cache
	Limit  int64
	Window time.Duration
}

func NewRateLimitService(cfg RateLimitConfig) *RateLimitConfig {
	return &RateLimitConfig{
		Store:  cfg.Store,
		Limit:  cfg.Limit,
		Window: cfg.Window,
	}
}

func (r *RateLimitConfig) GetLimit() int64 {
	return r.Limit
}

func (r *RateLimitConfig) RateLimit(key string) (count int64, remaining int64, resetIn time.Duration, err error) {
	count, resetIn, err = r.Store.Increment(key, r.Window)
	if err != nil {
		return 0, 0, 0, err
	}

	remaining = r.Limit - count
	if remaining < 0 {
		return count, 0, resetIn, ErrRateLimitExceeded
	}
	return count, remaining, resetIn, nil
}
