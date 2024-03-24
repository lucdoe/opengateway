package data

import (
	"context"
	"time"
)

type CacheClient interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
}

// implements CacheClient
type redisClient struct {
	client RedisClient
}

// abstracts the go-redis client
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

// when called: instanciates a redisClient
func NewRedisCache(client RedisClient) CacheClient {
	return &redisClient{client: client}
}

func (r *redisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(context.Background(), key, value, expiration)
}

func (r *redisClient) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key)
}
