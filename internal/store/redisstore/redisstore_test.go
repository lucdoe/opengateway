// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package redisstore_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/lucdoe/opengateway/internal/store"
	"github.com/lucdoe/opengateway/internal/store/redisstore"
	"github.com/lucdoe/opengateway/internal/store/storetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	incrementShouldNotErrorMsg    = "Increment should not error"
	incrementShouldReturnIncValue = "Increment should return incremented value"
)

func TestStoreOperations(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cacheConfig := redisstore.Config{
		Addr:     mr.Addr(),
		Password: "",
	}
	redisCache, err := redisstore.New(cacheConfig)
	require.NoError(t, err)

	t.Run("Set and Get operations", func(t *testing.T) {
		keyValue, value := "testkey", "testvalue"
		setErr := redisCache.Set(keyValue, value, 10*time.Minute)
		assert.NoError(t, setErr, "Set should not error")

		retrievedValue, getErr := redisCache.Get(keyValue)
		assert.NoError(t, getErr, "Get should not error")
		assert.Equal(t, value, retrievedValue, "Get should retrieve what was set")
	})

	t.Run("Increment operation", func(t *testing.T) {
		incrementKey := "incrementkey"
		count, _, incErr := redisCache.Increment(incrementKey, 1*time.Minute)
		assert.NoError(t, incErr, incrementShouldNotErrorMsg)
		assert.Equal(t, int64(1), count, incrementShouldReturnIncValue)

		count, _, incErr = redisCache.Increment(incrementKey, 1*time.Minute)
		assert.NoError(t, incErr, incrementShouldNotErrorMsg)
		assert.Equal(t, int64(2), count, incrementShouldReturnIncValue)
	})

	t.Run("Expiration handling", func(t *testing.T) {
		expireKey := "expirekey"
		value := "expirevalue"
		setErr := redisCache.Set(expireKey, value, 1*time.Second)
		assert.NoError(t, setErr, "Set should not error")

		retrievedValue, getErr := redisCache.Get(expireKey)
		assert.NoError(t, getErr, "Get should not error")
		assert.Equal(t, value, retrievedValue, "Get should retrieve what was set")

		mr.FastForward(1 * time.Second)

		retrievedValue, getErr = redisCache.Get(expireKey)
		assert.ErrorIs(t, getErr, store.ErrNotFound, "Get reports an expired key as not found")
		assert.Equal(t, "", retrievedValue, "Get should return an empty string after expiration")
	})

	t.Run("Increment and Expiration handling", func(t *testing.T) {
		incrementExpireKey := "incrementexpirekey"
		count, _, incErr := redisCache.Increment(incrementExpireKey, 1*time.Second)
		assert.NoError(t, incErr, incrementShouldNotErrorMsg)
		assert.Equal(t, int64(1), count, incrementShouldReturnIncValue)

		mr.FastForward(1 * time.Second)

		count, _, incErr = redisCache.Increment(incrementExpireKey, 1*time.Second)
		assert.NoError(t, incErr, "Increment should not error after expiration")
		assert.Equal(t, int64(1), count, "Increment should reset after expiration")
	})

	t.Run("Increment reports the time left in the window", func(t *testing.T) {
		key := "resetkey"
		_, resetIn, incErr := redisCache.Increment(key, time.Minute)
		assert.NoError(t, incErr)
		assert.Equal(t, time.Minute, resetIn)

		mr.FastForward(20 * time.Second)
		_, resetIn, incErr = redisCache.Increment(key, time.Minute)
		assert.NoError(t, incErr)
		assert.Equal(t, 40*time.Second, resetIn, "later increments do not extend the window")
	})

	t.Run("Increment gives a counter without expiry one", func(t *testing.T) {
		key := "stuckkey"
		require.NoError(t, mr.Set(key, "41")) // left behind with no expiry

		count, resetIn, incErr := redisCache.Increment(key, time.Minute)

		assert.NoError(t, incErr)
		assert.Equal(t, int64(42), count)
		assert.Equal(t, time.Minute, resetIn)
		assert.Equal(t, time.Minute, mr.TTL(key))
	})

	t.Run("Error handling on Set", func(t *testing.T) {
		mr.Close()
		setErr := redisCache.Set("errorKey", "value", 10*time.Minute)
		assert.Error(t, setErr, "Set should error when Redis is unavailable")
	})

	t.Run("Error handling on Get", func(t *testing.T) {
		_, getErr := redisCache.Get("errorKey")
		assert.Error(t, getErr, "Get should error when Redis is unavailable")
	})

	t.Run("Error handling on Increment", func(t *testing.T) {
		_, _, incErr := redisCache.Increment("errorKey", 10*time.Minute)
		assert.Error(t, incErr, "Increment should error when Redis is unavailable")
	})
}

func TestNewFailsWhenRedisIsUnreachable(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	addr := mr.Addr()
	mr.Close()

	redisCache, err := redisstore.New(redisstore.Config{Addr: addr})

	assert.Nil(t, redisCache)
	assert.ErrorContains(t, err, "connecting to redis at "+addr)
}

func TestStoreConformance(t *testing.T) {
	storetest.Run(t, func(t *testing.T) storetest.Backend {
		mr := miniredis.RunT(t)
		s, err := redisstore.New(redisstore.Config{Addr: mr.Addr()})
		require.NoError(t, err)
		return storetest.Backend{Store: s, Advance: mr.FastForward}
	})
}
