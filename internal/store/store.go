// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

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
