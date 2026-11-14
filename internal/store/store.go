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

// Package store defines the key-value store the cache and rate-limit plugins
// keep their state in. Backends live in the sub-packages: memory keeps it in
// the gateway process, redisstore in a Redis-compatible server. Another
// backend only has to implement Store.
package store

import (
	"errors"
	"time"
)

// ErrNotFound is returned by Get for a key that does not exist or expired.
var ErrNotFound = errors.New("store: key not found")

// Store keeps string values and counters, each with an expiry. It must be
// safe for concurrent use.
type Store interface {
	// Get returns the value at key, or ErrNotFound.
	Get(key string) (string, error)
	// Set stores value at key for ttl.
	Set(key, value string, ttl time.Duration) error
	// Increment adds one to the counter at key and returns the new count and
	// the time left until it resets. A counter starts at 1 and lives for
	// window from its first increment; later increments do not extend it.
	Increment(key string, window time.Duration) (count int64, resetIn time.Duration, err error)
}
