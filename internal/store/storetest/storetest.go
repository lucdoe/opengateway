// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package storetest checks that a store.Store backend behaves the way the
// cache and rate-limit plugins rely on. Every backend runs it from its tests.
package storetest

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Backend is a fresh, empty store and a way to move its clock forward.
type Backend struct {
	Store   store.Store
	Advance func(time.Duration)
}

// Run checks the backend that newBackend returns; it is called once per
// test case.
func Run(t *testing.T, newBackend func(t *testing.T) Backend) {
	t.Run("Get of a missing key is ErrNotFound", func(t *testing.T) {
		b := newBackend(t)
		_, err := b.Store.Get("missing")
		assert.ErrorIs(t, err, store.ErrNotFound)
	})

	t.Run("Set then Get returns the value until it expires", func(t *testing.T) {
		b := newBackend(t)
		require.NoError(t, b.Store.Set("k", "v", time.Minute))

		value, err := b.Store.Get("k")
		require.NoError(t, err)
		assert.Equal(t, "v", value)

		b.Advance(time.Minute)
		_, err = b.Store.Get("k")
		assert.ErrorIs(t, err, store.ErrNotFound)
	})

	t.Run("Set replaces the value and its expiry", func(t *testing.T) {
		b := newBackend(t)
		require.NoError(t, b.Store.Set("k", "old", time.Second))
		require.NoError(t, b.Store.Set("k", "new", time.Minute))

		b.Advance(30 * time.Second)
		value, err := b.Store.Get("k")
		require.NoError(t, err)
		assert.Equal(t, "new", value)
	})

	t.Run("Increment counts within one window", func(t *testing.T) {
		b := newBackend(t)
		for want := int64(1); want <= 3; want++ {
			count, resetIn, err := b.Store.Increment("c", time.Minute)
			require.NoError(t, err)
			assert.Equal(t, want, count)
			assert.Equal(t, time.Minute, resetIn)
		}
	})

	t.Run("Increment does not extend the window", func(t *testing.T) {
		b := newBackend(t)
		_, _, err := b.Store.Increment("c", time.Minute)
		require.NoError(t, err)

		b.Advance(20 * time.Second)
		count, resetIn, err := b.Store.Increment("c", time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
		assert.Equal(t, 40*time.Second, resetIn)
	})

	t.Run("Increment starts over after the window", func(t *testing.T) {
		b := newBackend(t)
		for i := 0; i < 5; i++ {
			_, _, err := b.Store.Increment("c", time.Minute)
			require.NoError(t, err)
		}

		b.Advance(time.Minute)
		count, resetIn, err := b.Store.Increment("c", time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, time.Minute, resetIn)
	})

	t.Run("counters are separate per key", func(t *testing.T) {
		b := newBackend(t)
		_, _, err := b.Store.Increment("a", time.Minute)
		require.NoError(t, err)
		count, _, err := b.Store.Increment("b", time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("concurrent increments are all counted", func(t *testing.T) {
		b := newBackend(t)
		const workers, each = 8, 50
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(w int) {
				defer wg.Done()
				for i := 0; i < each; i++ {
					_, _, err := b.Store.Increment("shared", time.Minute)
					assert.NoError(t, err, fmt.Sprintf("worker %d", w))
				}
			}(w)
		}
		wg.Wait()

		count, _, err := b.Store.Increment("shared", time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(workers*each+1), count)
	})
}
