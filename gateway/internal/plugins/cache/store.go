package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheStore interface {
	Increment(key string, window time.Duration) (int64, error)
	Expire(key string, window time.Duration) error
}

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) CacheStore {
	return &RedisStore{
		client: client,
	}
}

func (s *RedisStore) Increment(key string, window time.Duration) (int64, error) {
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

func (s *RedisStore) Expire(key string, window time.Duration) error {
	ctx := context.Background()
	_, err := s.client.Expire(ctx, key, window).Result()
	return err
}
