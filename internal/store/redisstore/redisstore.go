// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package redisstore is a store kept in a Redis server, or one that speaks
// the same protocol such as Valkey, KeyDB or Dragonfly. Several gateway
// instances pointed at the same server share rate limits and cache. It needs
// Redis 7.0 or newer.
package redisstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lucdoe/opengateway/internal/store"
	"github.com/redis/go-redis/v9"
)

// Config says where the Redis server is.
type Config struct {
	Addr     string
	Password string
}

// Store is a store.Store backed by Redis.
type Store struct {
	client *redis.Client
}

var _ store.Store = (*Store)(nil)

// New connects to the Redis server at cfg.Addr. It fails if the server does
// not answer.
func New(cfg Config) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		client.Close()
		return nil, fmt.Errorf("connecting to redis at %s: %w", cfg.Addr, err)
	}

	return &Store{client: client}, nil
}

// Get returns the value at key, or store.ErrNotFound.
func (r *Store) Get(key string) (string, error) {
	value, err := r.client.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return "", store.ErrNotFound
	}
	return value, err
}

// Set stores value at key for ttl.
func (r *Store) Set(key string, value string, ttl time.Duration) error {
	return r.client.Set(context.Background(), key, value, ttl).Err()
}

// Increment adds one to the counter at key and returns the new count and the
// time left until the counter resets. A counter lives for window from its
// first increment.
//
// The increment and the expiry run in one MULTI transaction, so no counter is
// ever left without an expiry. PEXPIRE NX only sets an expiry where there is
// none, which also repairs counters an older version left without one. NX
// needs Redis 7.0 or newer.
func (r *Store) Increment(key string, window time.Duration) (count int64, resetIn time.Duration, err error) {
	ctx := context.Background()
	var incr *redis.IntCmd
	var ttl *redis.DurationCmd
	_, err = r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		incr = pipe.Incr(ctx, key)
		pipe.Do(ctx, "PEXPIRE", key, window.Milliseconds(), "NX")
		ttl = pipe.PTTL(ctx, key)
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	return incr.Val(), ttl.Val(), nil
}
