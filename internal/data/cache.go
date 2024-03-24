package data

import (
	"context"
	"time"
)

type CacheClient interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
}

type concreteRedisClient struct {
	client RedisClient
}

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

func NewRedisCache(client RedisClient) CacheClient {
	return &concreteRedisClient{client: client}
}

func (r *concreteRedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(context.Background(), key, value, expiration)
}

func (r *concreteRedisClient) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key)
}
