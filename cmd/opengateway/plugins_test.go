// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"io"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/store"
	"github.com/lucdoe/opengateway/internal/store/memory"
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
	return func(config.StoreConfig) (store.Store, error) {
		t.Error("the store must not be connected to")
		return nil, errors.New("unexpected store connection")
	}
}

func workingStore(connections *int) storeFactory {
	return func(config.StoreConfig) (store.Store, error) {
		*connections++
		return memory.New(), nil
	}
}

func builtPlugins(t *testing.T, cfg *config.Config, newStore storeFactory) []string {
	t.Helper()
	plugins, closers, err := buildPlugins(cfg, newStore)
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, closer := range closers {
			closer.Close()
		}
	})

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

func TestBuildPluginsOpensTheConfiguredStore(t *testing.T) {
	cfg := configUsing("cache")
	cfg.Store = config.StoreConfig{Type: "redis", Redis: config.RedisConfig{Host: "redis.internal", Port: 6380, Password: "secret"}}

	var received config.StoreConfig
	newStore := func(s config.StoreConfig) (store.Store, error) {
		received = s
		return memory.New(), nil
	}

	builtPlugins(t, cfg, newStore)

	assert.Equal(t, cfg.Store, received)
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
	failingStore := func(config.StoreConfig) (store.Store, error) { return nil, errors.New("connection refused") }

	tests := []struct {
		name     string
		cfg      func(t *testing.T) *config.Config
		newStore storeFactory
		expected string
	}{
		{
			name:     "auth without a secret",
			cfg:      func(t *testing.T) *config.Config { return configUsing("auth") },
			expected: "auth plugin: needs a secret key or a JWKS URL",
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
			var closers []io.Closer
			assert.NotPanics(t, func() { plugins, closers, err = buildPlugins(tc.cfg(t), newStore) })

			assert.Nil(t, plugins)
			assert.Nil(t, closers, "what was built is closed, not handed back")
			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestBuildPluginsReturnsTheLoggerToClose(t *testing.T) {
	cfg := configUsing("logger")
	path := filepath.Join(t.TempDir(), "server.log")
	cfg.Plugins.LoggerConfig.FilePath = path

	_, closers, err := buildPlugins(cfg, noStore(t))
	require.NoError(t, err)
	require.Len(t, closers, 1)

	assert.NoError(t, closers[0].Close())
}

// The logger built first is closed when a later plugin fails.
func TestBuildPluginsClosesTheLoggerWhenALaterPluginFails(t *testing.T) {
	cfg := configUsing("logger", "rate-limit")
	cfg.Plugins.LoggerConfig.FilePath = filepath.Join(t.TempDir(), "server.log")
	cfg.Plugins.RateLimitConfig = config.RateLimitConfig{Limit: 10, Window: time.Minute}
	failingStore := func(config.StoreConfig) (store.Store, error) { return nil, errors.New("connection refused") }

	plugins, closers, err := buildPlugins(cfg, failingStore)

	assert.Nil(t, plugins)
	assert.Nil(t, closers)
	assert.ErrorContains(t, err, "rate-limit plugin: connection refused")
}

func TestRateLimitKeyFollowsTheConfig(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.5:1234"
	req.Header.Set("X-API-Key", "k-1")

	assert.Equal(t, "203.0.113.5", rateLimitKey("")(req))
	assert.Equal(t, "203.0.113.5", rateLimitKey("ip")(req))
	assert.Regexp(t, `^header:[0-9a-f]{32}$`, rateLimitKey("header:X-API-Key")(req))
	assert.Equal(t, "203.0.113.5", rateLimitKey("jwt:sub")(req), "before auth or without a token, the IP counts")
}
