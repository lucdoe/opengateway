// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type CacheConfig struct {
	Addr     string
	Password string
}

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	GenerateCacheKey(r *http.Request) string
	Increment(key string, window time.Duration) (int64, error)
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(cfg CacheConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
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
		if err != nil {
			return 0, err
		}
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
