// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/plugins/auth"
	"github.com/lucdoe/opengateway/internal/store"
	"github.com/lucdoe/opengateway/internal/store/memory"
	"github.com/lucdoe/opengateway/internal/store/redisstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loads returns a config loader that always returns cfg and err.
func loads(cfg *config.Config, err error) func(string) (*config.Config, error) {
	return func(string) (*config.Config, error) { return cfg, err }
}

func TestInitializeServerSuccess(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	newProxy := func(target string, timeout time.Duration) (http.Handler, error) {
		return http.NotFoundHandler(), nil
	}

	storeConnections := 0
	newStore := func(config.StoreConfig) (store.Store, error) {
		storeConnections++
		return memory.New(), nil
	}

	cfg := &config.Config{
		Services: map[string]config.Service{
			"starwars": {
				URL:     "https://swapi.dev/api",
				Plugins: []string{"rate-limit", "logger", "cache", "cors"},
				Endpoints: []config.Endpoint{
					{
						Name:       "List People",
						HTTPMethod: "GET",
						Path:       "/people",
					},
				},
			},
		},
		Plugins: config.PluginConfig{
			LoggerConfig: config.LoggerConfig{
				FilePath: filepath.Join(t.TempDir(), "server.log"),
			},
			JWTConfig: config.JWTConfig{
				SecretKey:     "your-256-bit-secret",
				SigningMethod: "HS256",
				Issuer:        "ExampleIssuer",
				Audience:      "ExampleAudience",
				Scope:         "ExampleScope",
			},
			RateLimitConfig: config.RateLimitConfig{
				Limit:  100,
				Window: 60 * time.Second,
			},
			CORSConfig: config.CORSConfig{
				Origins: "*",
				Methods: "GET, POST, PUT, DELETE, OPTIONS",
				Headers: "Content-Type, Authorization",
			},
		},
	}

	var loadedPath string
	deps := ServerDependencies{
		loadConfig: func(path string) (*config.Config, error) {
			loadedPath = path
			return cfg, nil
		},
		router:   mux.NewRouter(),
		newProxy: newProxy,
		newStore: newStore,
	}

	server, err := InitializeServer(deps)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, 1, storeConnections, "cache and rate-limit must share one store connection")
	assert.Equal(t, "./cmd/opengateway/config.yaml", loadedPath, "the default path without CONFIG_PATH")
}

func TestInitializeServerReturnsTheConfigError(t *testing.T) {
	server, err := InitializeServer(ServerDependencies{loadConfig: loads(nil, errors.New("invalid configuration")), router: mux.NewRouter()})

	assert.Nil(t, server)
	assert.ErrorContains(t, err, "invalid configuration")
}

func TestInitializeServerReturnsThePluginError(t *testing.T) {
	cfg := configUsing("cache")
	server, err := InitializeServer(ServerDependencies{
		loadConfig: loads(cfg, nil),
		router:     mux.NewRouter(),
		newProxy:   func(string, time.Duration) (http.Handler, error) { return http.NotFoundHandler(), nil },
		newStore:   func(config.StoreConfig) (store.Store, error) { return nil, errors.New("connection refused") },
	})

	assert.Nil(t, server)
	assert.ErrorContains(t, err, "cache plugin: connection refused")
}

func TestOpenStoreOpensTheConfiguredType(t *testing.T) {
	t.Run("memory", func(t *testing.T) {
		s, err := openStore(config.StoreConfig{Type: config.StoreMemory})
		require.NoError(t, err)
		assert.IsType(t, &memory.Store{}, s)
	})

	t.Run("redis", func(t *testing.T) {
		mr := miniredis.RunT(t)
		host, port, _ := net.SplitHostPort(mr.Addr())
		portNumber, _ := strconv.Atoi(port)

		s, err := openStore(config.StoreConfig{Type: config.StoreRedis, Redis: config.RedisConfig{Host: host, Port: portNumber}})
		require.NoError(t, err)
		assert.IsType(t, &redisstore.Store{}, s)
	})

	t.Run("redis that is not reachable", func(t *testing.T) {
		s, err := openStore(config.StoreConfig{Type: config.StoreRedis, Redis: config.RedisConfig{Host: "127.0.0.1", Port: 1}})
		assert.Nil(t, s)
		assert.ErrorContains(t, err, "connecting to redis at 127.0.0.1:1")
	})

	t.Run("unknown", func(t *testing.T) {
		s, err := openStore(config.StoreConfig{Type: "memcached"})
		assert.Nil(t, s)
		assert.ErrorContains(t, err, `unknown store type "memcached"`)
	})
}

// The real auth plugin and scope check, wired the way main wires them.
func TestInitializeServerEnforcesTokenClaims(t *testing.T) {
	cfg, err := config.Parse([]byte(`
Plugins:
  JWTConfig:
    SecretKey: secret
    Issuer: issuer
    Audience: gateway
    Scope: read
Services:
  api:
    URL: http://upstream.test
    Endpoints:
      - Name: Items
        HTTPMethod: GET
        Path: /items
        Auth: {ApplyAuth: true}
      - Name: Admin
        HTTPMethod: GET
        Path: /admin
        Auth: {ApplyAuth: true, Scope: admin}
`))
	require.NoError(t, err)
	s, err := InitializeServer(ServerDependencies{
		loadConfig: loads(cfg, nil),
		router:     mux.NewRouter(),
		newProxy: func(string, time.Duration) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), nil
		},
	})
	require.NoError(t, err)

	token := func(issuer, scope string) string {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{Issuer: issuer, Audience: []string{"gateway"}},
			Scope:            scope,
		}).SignedString([]byte("secret"))
		require.NoError(t, err)
		return signed
	}
	tests := []struct {
		name, path, token string
		status            int
	}{
		{"valid token", "/items", token("issuer", "read"), http.StatusOK},
		{"wrong issuer", "/items", token("someone-else", "read"), http.StatusUnauthorized},
		{"lacks the global scope", "/items", token("issuer", "admin"), http.StatusForbidden},
		{"lacks the endpoint scope", "/admin", token("issuer", "read"), http.StatusForbidden},
		{"has both scopes", "/admin", token("issuer", "read admin"), http.StatusOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			rec := httptest.NewRecorder()

			s.Router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
		})
	}
}

// main's own wiring must start the gateway from the shipped config.
func TestDefaultDependenciesStartTheShippedConfig(t *testing.T) {
	t.Setenv("CONFIG_PATH", "config.yaml")

	s, err := InitializeServer(defaultDependencies())
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, closer := range s.Closers {
			closer.Close()
		}
	})

	assert.Equal(t, ":4000", s.Addr)
	assert.NotEmpty(t, s.Closers, "the shipped config uses the logger")
}
