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

type Cache interface {
	Increment(key string, window time.Duration) (int64, error)
}

type RateLimiter interface {
	RateLimit(key string) (count int64, remaining int64, window time.Duration, err error)
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

func (r *RateLimitConfig) RateLimit(key string) (count int64, remaining int64, window time.Duration, err error) {
	curCount, err := r.Store.Increment(key, r.Window)
	if err != nil {
		return 0, 0, 0, err
	}

	curRemaining := r.Limit - curCount
	if curRemaining < 0 {
		return curCount, 0, 0, ErrRateLimitExceeded
	}

	return curCount, curRemaining, r.Window, nil
}
