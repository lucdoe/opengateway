package cache_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
	"github.com/stretchr/testify/assert"
)

func TestIncrementAndExpire(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	cacheMiddleware := cache.NewCacheMiddleware(mr.Addr(), "")

	key := "testkey"
	window := 1 * time.Minute
	count, err := cacheMiddleware.Increment(key, window)
	assert.NoError(t, err, "Increment should not error")
	assert.Equal(t, int64(1), count, "Increment should return 1 on first call")

	time.Sleep(window + 1*time.Second)
	exists := mr.Exists(key)
	assert.False(t, exists, "Key should not exist after expiration")
}

func TestIncrementFailure(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	mr.Close()

	cacheMiddleware := cache.NewCacheMiddleware(mr.Addr(), "")

	_, err = cacheMiddleware.Increment("shouldFail", 1*time.Minute)
	assert.Error(t, err, "Increment should error when Redis is down")
}
