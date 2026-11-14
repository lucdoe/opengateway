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

var errLimitExceeded = errors.New("rate limit exceeded")

// Counter is the part of a store.Store the counters live in.
type Counter interface {
	// Increment adds one to the counter at key and returns the new count and
	// the time left until it resets. A counter lives for window from its
	// first increment.
	Increment(key string, window time.Duration) (count int64, resetIn time.Duration, err error)
}

// Limiter allows limit requests per key per window.
type Limiter struct {
	counters Counter
	limit    int64
	window   time.Duration
}

func NewLimiter(counters Counter, limit int64, window time.Duration) *Limiter {
	return &Limiter{counters: counters, limit: limit, window: window}
}

// Take counts one request for key. remaining is how many more are allowed
// and resetIn how long until the count starts over. It returns
// errLimitExceeded once the limit is used up, and the store's error if
// counting failed.
func (l *Limiter) Take(key string) (remaining int64, resetIn time.Duration, err error) {
	count, resetIn, err := l.counters.Increment(key, l.window)
	if err != nil {
		return 0, 0, err
	}
	if count > l.limit {
		return 0, resetIn, errLimitExceeded
	}
	return l.limit - count, resetIn, nil
}
