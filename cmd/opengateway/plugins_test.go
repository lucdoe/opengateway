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

package main

import (
	"errors"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// configUsing returns a config with one service that lists the given plugins.
// It carries no plugin settings.
func configUsing(plugins ...string) *config.Config {
	return &config.Config{
		Services: map[string]config.Service{
			"api": {
				URL:       "https://example.com",
				Plugins:   plugins,
				Endpoints: []config.Endpoint{{Name: "Items", HTTPMethod: "GET", Path: "/items"}},
			},
		},
	}
}

// noStore fails the test if a plugin asks for the store.
func noStore(t *testing.T) storeFactory {
	return func(config.RedisConfig) (cache.Cache, error) {
		t.Error("the store must not be connected to")
		return nil, errors.New("unexpected store connection")
	}
}

func workingStore(connections *int) storeFactory {
	return func(config.RedisConfig) (cache.Cache, error) {
		*connections++
		return new(MockCacheService), nil
	}
}

func builtPlugins(t *testing.T, cfg *config.Config, newStore storeFactory) []string {
	t.Helper()
	plugins, err := buildPlugins(cfg, newStore)
	require.NoError(t, err)

	built := make([]string, 0, len(plugins))
	for name, plugin := range plugins {
		assert.NotNil(t, plugin, name)
		built = append(built, name)
	}
	sort.Strings(built)
	return built
}

func TestBuildPluginsBuildsOnlyWhatIsUsed(t *testing.T) {
	t.Run("nothing used", func(t *testing.T) {
		assert.Empty(t, builtPlugins(t, configUsing(), noStore(t)))
	})

	t.Run("cors only needs no settings and no store", func(t *testing.T) {
		assert.Equal(t, []string{"cors"}, builtPlugins(t, configUsing("cors"), noStore(t)))
	})

	t.Run("logger only", func(t *testing.T) {
		cfg := configUsing("logger")
		cfg.Plugins.LoggerConfig.FilePath = filepath.Join(t.TempDir(), "server.log")

		assert.Equal(t, []string{"logger"}, builtPlugins(t, cfg, noStore(t)))
	})

	t.Run("auth through ApplyAuth", func(t *testing.T) {
		cfg := configUsing()
		cfg.Services["api"].Endpoints[0].Auth.ApplyAuth = true
		cfg.Plugins.JWTConfig.SecretKey = "secret"

		assert.Equal(t, []string{"auth"}, builtPlugins(t, cfg, noStore(t)))
	})

	t.Run("everything", func(t *testing.T) {
		connections := 0
		cfg := configUsing("logger", "cors", "rate-limit", "auth", "cache")
		cfg.Plugins.LoggerConfig.FilePath = filepath.Join(t.TempDir(), "server.log")
		cfg.Plugins.JWTConfig.SecretKey = "secret"
		cfg.Plugins.RateLimitConfig = config.RateLimitConfig{Limit: 10, Window: time.Minute}

		built := builtPlugins(t, cfg, workingStore(&connections))

		assert.Equal(t, []string{"auth", "cache", "cors", "logger", "rate-limit"}, built)
		assert.Equal(t, 1, connections, "cache and rate-limit must share one store connection")
	})
}

func TestBuildPluginsConnectsToTheConfiguredRedis(t *testing.T) {
	cfg := configUsing("cache")
	cfg.Redis = config.RedisConfig{Host: "redis.internal", Port: 6380, Password: "secret"}

	var received config.RedisConfig
	newStore := func(redis config.RedisConfig) (cache.Cache, error) {
		received = redis
		return new(MockCacheService), nil
	}

	builtPlugins(t, cfg, newStore)

	assert.Equal(t, cfg.Redis, received)
}

// Before, the gateway refused to start without a JWT secret even when no
// endpoint used auth.
func TestBuildPluginsDoesNotNeedSettingsOfUnusedPlugins(t *testing.T) {
	connections := 0
	cfg := configUsing("cache")

	assert.Equal(t, []string{"cache"}, builtPlugins(t, cfg, workingStore(&connections)))
	assert.Equal(t, 1, connections)
}

func TestBuildPluginsReturnsErrorsInsteadOfPanicking(t *testing.T) {
	failingStore := func(config.RedisConfig) (cache.Cache, error) { return nil, errors.New("connection refused") }

	tests := []struct {
		name     string
		cfg      func(t *testing.T) *config.Config
		newStore storeFactory
		expected string
	}{
		{
			name:     "auth without a secret",
			cfg:      func(t *testing.T) *config.Config { return configUsing("auth") },
			expected: "auth plugin: secret key must not be empty",
		},
		{
			name: "log file cannot be opened",
			cfg: func(t *testing.T) *config.Config {
				cfg := configUsing("logger")
				cfg.Plugins.LoggerConfig.FilePath = filepath.Join(t.TempDir(), "missing-dir", "server.log")
				return cfg
			},
			expected: "logger plugin: opening log file",
		},
		{
			name:     "store unreachable for cache",
			cfg:      func(t *testing.T) *config.Config { return configUsing("cache") },
			newStore: failingStore,
			expected: "cache plugin: connection refused",
		},
		{
			name:     "store unreachable for rate-limit",
			cfg:      func(t *testing.T) *config.Config { return configUsing("rate-limit") },
			newStore: failingStore,
			expected: "rate-limit plugin: connection refused",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newStore := tc.newStore
			if newStore == nil {
				newStore = noStore(t)
			}

			var plugins any
			var err error
			assert.NotPanics(t, func() { plugins, err = buildPlugins(tc.cfg(t), newStore) })

			assert.Nil(t, plugins)
			assert.ErrorContains(t, err, tc.expected)
		})
	}
}
