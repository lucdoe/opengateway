package cache

import (
	"context"
	"net/http"
	"time"

	"github.com/lucdoe/open-gateway/gateway/internal"
	"github.com/redis/go-redis/v9"
)

type CacheMiddleware interface {
	Middleware(next http.Handler) http.Handler
	Increment(key string, window time.Duration) (int64, error)
}

type cacheStore struct {
	client *redis.Client
}

func NewCacheMiddleware(addr string, password string) CacheMiddleware {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}

	return &cacheStore{client: client}
}

func (s *cacheStore) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheKey := generateCacheKey(r)
		cachedResponse, err := s.get(cacheKey)
		if err == nil && cachedResponse != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cachedResponse))
			return
		}

		recorder := internal.NewResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if recorder.StatusCode == http.StatusOK {
			responseBody := recorder.Body.String()
			s.set(cacheKey, responseBody, 10*time.Minute)
		}

		recorder.CopyBody(w)
	})
}

func (s *cacheStore) Increment(key string, window time.Duration) (int64, error) {
	ctx := context.Background()
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	var isFirstRun = count == 1
	if isFirstRun {
		_, err = s.client.Expire(ctx, key, window).Result()
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func (s *cacheStore) get(key string) (string, error) {
	result, err := s.client.Get(context.Background(), key).Result()
	return result, err
}

func (s *cacheStore) set(key string, value string, window time.Duration) error {
	_, err := s.client.Set(context.Background(), key, value, window).Result()
	return err
}
