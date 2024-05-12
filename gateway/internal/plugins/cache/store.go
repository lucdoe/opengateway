package cache

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	GenerateCacheKey(r *http.Request) string
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string, password string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}

	return &RedisCache{client: client}
}

func (r *RedisCache) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
}

func (r *RedisCache) Set(key string, value string, expiration time.Duration) error {
	_, err := r.client.Set(context.Background(), key, value, expiration).Result()
	return err
}

func (r *RedisCache) Increment(key string, window time.Duration) (int64, error) {
	ctx := context.Background()
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		_, err := r.client.Expire(ctx, key, window).Result()
		return count, err
	}

	return count, nil
}

func (r *RedisCache) GenerateCacheKey(rq *http.Request) string {
	path := rq.URL.Path
	address := rq.RemoteAddr
	return fmt.Sprintf("%s:%s:%s:%s", rq.Method, path, r.sortQueryParams(rq.URL.Query()), address)
}

func (r *RedisCache) sortQueryParams(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sortedParams []string
	for _, k := range keys {
		sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", k, params.Get(k)))
	}
	return strings.Join(sortedParams, "&")
}
