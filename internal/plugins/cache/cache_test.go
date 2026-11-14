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

package cache_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	incrementShouldNotErrorMsg    = "Increment should not error"
	incrementShouldReturnIncValue = "Increment should return incremented value"
)

func TestRedisCacheOperations(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cacheConfig := cache.CacheConfig{
		Addr:     mr.Addr(),
		Password: "",
	}
	redisCache, err := cache.NewRedisCache(cacheConfig)
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
		assert.Error(t, getErr, "Get should error after expiration")
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

func TestNewRedisCacheFailsWhenRedisIsUnreachable(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	addr := mr.Addr()
	mr.Close()

	redisCache, err := cache.NewRedisCache(cache.CacheConfig{Addr: addr})

	assert.Nil(t, redisCache)
	assert.ErrorContains(t, err, "connecting to redis at "+addr)
}
