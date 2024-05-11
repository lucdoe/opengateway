package cache

import (
	"context"
	"net/http"
	"time"

	server "github.com/lucdoe/open-gateway/gateway/internal/server"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Init() error
	Increment(key string, window time.Duration) (int64, error)
	Expire(key string, window time.Duration) error
	Apply(next http.Handler) http.Handler
	Configure(settings map[string]interface{}) error
}

type CacheStore struct {
	client *redis.Client
}

func NewRedisStore() Cache {
	return &CacheStore{}
}

func (s *CacheStore) Init() error {
	s.client = redis.NewClient(&redis.Options{
		Addr:     "",
		Password: "",
	})
	if _, err := s.client.Ping(context.Background()).Result(); err != nil {
		return err
	}
	return nil
}

func (s *CacheStore) Get(key string) (string, error) {
	ctx := context.Background()
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *CacheStore) Set(key string, value string, window time.Duration) error {
	ctx := context.Background()
	_, err := s.client.Set(ctx, key, value, window).Result()
	return err
}

func (s *CacheStore) Increment(key string, window time.Duration) (int64, error) {
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

func (s *CacheStore) Expire(key string, window time.Duration) error {
	ctx := context.Background()
	_, err := s.client.Expire(ctx, key, window).Result()
	return err
}

func (s *CacheStore) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheKey := generateCacheKey(r)

		cachedResponse, err := s.Get(cacheKey)
		if err == nil && cachedResponse != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cachedResponse))
			return
		}

		recorder := server.NewResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if recorder.StatusCode == http.StatusOK {
			responseBody := recorder.Body.String()
			s.Set(cacheKey, responseBody, 10*time.Minute)
		}

		recorder.CopyBody(w)
	})
}

func (s *CacheStore) Configure(settings map[string]interface{}) error {
	return nil
}
