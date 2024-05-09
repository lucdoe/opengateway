package cache_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	store "github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestIncrementAndExpire(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	store := store.NewRedisStore(client)

	// Test increment
	key := "testkey"
	window := 1 * time.Minute
	count, err := store.Increment(key, window)
	assert.NoError(t, err, "Increment should not error")
	assert.Equal(t, int64(1), count, "Increment should return 1 on first call")

	// Test if the key expires as expected
	time.Sleep(2 * time.Second)
	exists := mr.Exists(key)
	assert.True(t, exists, "Key should still exist")

	// Test expiration
	err = store.Expire(key, window)
	assert.NoError(t, err, "Expire should not error")

	// Check the key's TTL
	ttl := mr.TTL(key)
	assert.LessOrEqual(t, int(ttl/time.Second), int(window.Seconds()), "TTL should be less than or equal to the window")
}

func TestIncrementFailure(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := store.NewRedisStore(client)

	mr.Close()

	_, err = store.Increment("shouldFail", 1*time.Minute)
	assert.Error(t, err, "Increment should error when Redis is down")
}
