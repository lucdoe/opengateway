package cache_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
	"github.com/stretchr/testify/assert"
)

func TestRedisCacheOperations(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	cacheConfig := cache.CacheConfig{
		Addr:     mr.Addr(),
		Password: "",
	}
	redisCache := cache.NewRedisCache(cacheConfig)

	keyValue, value := "testkey", "testvalue"
	setErr := redisCache.Set(keyValue, value, 10*time.Minute)
	assert.NoError(t, setErr, "Set should not error")

	retrievedValue, getErr := redisCache.Get(keyValue)
	assert.NoError(t, getErr, "Get should not error")
	assert.Equal(t, value, retrievedValue, "Get should retrieve what was set")

	incrementKey := "incrementkey"
	count, incErr := redisCache.Increment(incrementKey, 1*time.Minute)
	assert.NoError(t, incErr, "Increment should not error")
	assert.Equal(t, int64(1), count, "Increment should return incremented value")

	req, _ := http.NewRequest("GET", "/path?b=2&a=1", nil)
	req.RemoteAddr = "123.45.67.89"
	expectedKey := "GET:/path:a=1&b=2:123.45.67.89"
	generatedKey := redisCache.GenerateCacheKey(req)
	assert.Equal(t, expectedKey, generatedKey, "GenerateCacheKey should return a correctly formatted key")
}
