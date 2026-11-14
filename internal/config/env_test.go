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

package config_test

import (
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalConfig = `
Services:
  api:
    URL: "https://example.com"
    Endpoints:
      - Name: "Items"
        HTTPMethod: "GET"
        Path: "/items"
`

const explicitSettings = `
Server:
  Port: 8080
  ReadHeaderTimeout: 5s
  IdleTimeout: 1m
  ShutdownTimeout: 20s
  UpstreamTimeout: 500ms
  CircuitBreaker:
    ConsecutiveFailures: 3
    OpenTimeout: 10s
    HalfOpenRequests: 2
Store:
  Type: "redis"
  Redis:
    Host: "redis.internal"
    Port: 6380
    Password: "from-file"
Plugins:
  JWTConfig:
    SecretKey: "from-file"
` + minimalConfig

// clearEnvironment makes the test independent of the variables set on the
// machine that runs it.
func clearEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{config.EnvPort, config.EnvStoreType, config.EnvRedisHost, config.EnvRedisPort, config.EnvRedisPassword, config.EnvJWTSecret} {
		t.Setenv(name, "")
	}
}

func TestParseAppliesDefaults(t *testing.T) {
	clearEnvironment(t)

	cfg, err := config.Parse([]byte(minimalConfig))
	require.NoError(t, err)

	assert.Equal(t, config.ServerConfig{
		Port:              4000,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ShutdownTimeout:   15 * time.Second,
		UpstreamTimeout:   30 * time.Second,
		CircuitBreaker:    config.CircuitBreakerConfig{ConsecutiveFailures: 5, OpenTimeout: 30 * time.Second, HalfOpenRequests: 1},
	}, cfg.Server)
	assert.Equal(t, config.StoreConfig{Type: "memory", Redis: config.RedisConfig{Host: "localhost", Port: 6379}}, cfg.Store)
	assert.Equal(t, ":4000", cfg.ListenAddr())
	assert.Equal(t, "localhost:6379", cfg.Store.Redis.Addr())
	assert.Empty(t, cfg.Plugins.JWTConfig.SecretKey)
}

func TestParseUsesTheFileOverDefaults(t *testing.T) {
	clearEnvironment(t)

	cfg, err := config.Parse([]byte(explicitSettings))
	require.NoError(t, err)

	assert.Equal(t, config.ServerConfig{
		Port:              8080,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
		ShutdownTimeout:   20 * time.Second,
		UpstreamTimeout:   500 * time.Millisecond,
		CircuitBreaker:    config.CircuitBreakerConfig{ConsecutiveFailures: 3, OpenTimeout: 10 * time.Second, HalfOpenRequests: 2},
	}, cfg.Server)
	assert.Equal(t, ":8080", cfg.ListenAddr())
	assert.Equal(t, "redis.internal:6380", cfg.Store.Redis.Addr())
	assert.Equal(t, "from-file", cfg.Store.Redis.Password)
	assert.Equal(t, "from-file", cfg.Plugins.JWTConfig.SecretKey)
}

func TestParseUsesTheEnvironmentOverTheFile(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("PORT", "9090")
	t.Setenv("REDIS_HOST", "redis.env")
	t.Setenv("REDIS_PORT", "6381")
	t.Setenv("REDIS_PASSWORD", "from-env")
	t.Setenv("JWT_SECRET", "from-env")

	cfg, err := config.Parse([]byte(explicitSettings))
	require.NoError(t, err)

	assert.Equal(t, ":9090", cfg.ListenAddr())
	assert.Equal(t, "redis.env:6381", cfg.Store.Redis.Addr())
	assert.Equal(t, "from-env", cfg.Store.Redis.Password)
	assert.Equal(t, "from-env", cfg.Plugins.JWTConfig.SecretKey)
}

func TestParseAppliesEachVariableOnItsOwn(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("REDIS_HOST", "redis")

	cfg, err := config.Parse([]byte(minimalConfig))
	require.NoError(t, err)

	assert.Equal(t, "redis:6379", cfg.Store.Redis.Addr(), "the host comes from the environment, the port from the default")
	assert.Equal(t, ":4000", cfg.ListenAddr())
}

func TestParseFormatsIPv6RedisHosts(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("REDIS_HOST", "::1")

	cfg, err := config.Parse([]byte(minimalConfig))
	require.NoError(t, err)

	assert.Equal(t, "[::1]:6379", cfg.Store.Redis.Addr())
}

// The secret can be left out of the file when the environment provides it.
func TestParseAcceptsTheJWTSecretFromTheEnvironmentOnly(t *testing.T) {
	withAuth := `
Services:
  api:
    URL: "https://example.com"
    Endpoints:
      - Name: "Items"
        HTTPMethod: "GET"
        Path: "/items"
        Auth:
          ApplyAuth: true
`
	clearEnvironment(t)

	_, err := config.Parse([]byte(withAuth))
	assert.ErrorContains(t, err, "Plugins.JWTConfig.SecretKey is required")

	t.Setenv("JWT_SECRET", "from-env")
	cfg, err := config.Parse([]byte(withAuth))
	require.NoError(t, err)
	assert.Equal(t, "from-env", cfg.Plugins.JWTConfig.SecretKey)
}

func TestParseRejectsInvalidTimeouts(t *testing.T) {
	clearEnvironment(t)

	tests := []struct {
		name     string
		server   string
		expected string
	}{
		{"negative timeout", "Server:\n  UpstreamTimeout: -5s\n", "Server.UpstreamTimeout -5s must be greater than 0"},
		{"not a duration", "Server:\n  IdleTimeout: soon\n", "cannot unmarshal"},
		{"negative breaker threshold", "Server:\n  CircuitBreaker:\n    ConsecutiveFailures: -1\n", "Server.CircuitBreaker.ConsecutiveFailures -1 must be greater than 0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Parse([]byte(tc.server + minimalConfig))

			assert.Nil(t, cfg)
			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestParseRejectsInvalidEnvironmentValues(t *testing.T) {
	tests := []struct {
		name     string
		variable string
		value    string
		expected string
	}{
		{"PORT is not a number", "PORT", "http", `environment variable PORT: "http" is not a number`},
		{"REDIS_PORT is not a number", "REDIS_PORT", "6379a", `environment variable REDIS_PORT: "6379a" is not a number`},
		{"PORT out of range", "PORT", "70000", "Server.Port 70000 must be between 1 and 65535"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearEnvironment(t)
			t.Setenv(tc.variable, tc.value)

			cfg, err := config.Parse([]byte(minimalConfig))

			assert.Nil(t, cfg)
			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestParseTakesTheStoreTypeFromTheEnvironment(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("STORE_TYPE", "redis")

	cfg, err := config.Parse([]byte(minimalConfig))
	require.NoError(t, err)

	assert.Equal(t, "redis", cfg.Store.Type)
}

func TestParseAcceptsTheStoreTypeInAnyCase(t *testing.T) {
	clearEnvironment(t)

	cfg, err := config.Parse([]byte("Store:\n  Type: \" Redis \"\n" + minimalConfig))
	require.NoError(t, err)

	assert.Equal(t, "redis", cfg.Store.Type)
}

func TestParseNamesTheNewPlaceOfTheRedisSection(t *testing.T) {
	clearEnvironment(t)

	cfg, err := config.Parse([]byte("Redis:\n  Host: \"redis\"\n" + minimalConfig))

	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "line 1: the top-level Redis section has moved: set Store.Type: redis and put its settings under Store.Redis")
}
