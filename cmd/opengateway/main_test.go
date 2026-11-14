// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
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
	"github.com/lucdoe/opengateway/internal/proxy"
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

	server, err := InitializeServer(deps, resolveConfigPath(""))
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, 1, storeConnections, "cache and rate-limit must share one store connection")
	assert.Equal(t, "./cmd/opengateway/config.yaml", loadedPath, "the default path without CONFIG_PATH")
}

func TestInitializeServerReturnsTheConfigError(t *testing.T) {
	server, err := InitializeServer(ServerDependencies{loadConfig: loads(nil, errors.New("invalid configuration")), router: mux.NewRouter()}, "config.yaml")

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
	}, "config.yaml")

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
	}, "config.yaml")
	require.NoError(t, err)

	token := func(issuer, scope string) string {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    issuer,
				Audience:  []string{"gateway"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
			Scope: scope,
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

// A response the upstream marks public is shared between callers of an
// authenticated endpoint, but only after auth and the scope check let them
// through: the cache runs after both.
func TestInitializeServerOnlyServesCachedAuthorizedResponsesToAllowedCallers(t *testing.T) {
	cfg, err := config.Parse([]byte(`
Plugins: {JWTConfig: {SecretKey: secret}}
Services:
  api:
    URL: http://upstream.test
    Plugins: [cache]
    Endpoints: [{HTTPMethod: GET, Path: /catalog, Auth: {ApplyAuth: true, Scope: read}}]
`))
	require.NoError(t, err)
	calls := 0
	s, err := InitializeServer(ServerDependencies{
		loadConfig: loads(cfg, nil),
		router:     mux.NewRouter(),
		newProxy: func(string, time.Duration) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Cache-Control", "public, max-age=60")
				w.Write([]byte("catalog"))
			}), nil
		},
		newStore: func(config.StoreConfig) (store.Store, error) { return memory.New(), nil },
	}, "config.yaml")
	require.NoError(t, err)

	token := func(subject, scope string) string {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{Subject: subject, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))},
			Scope:            scope,
		}).SignedString([]byte("secret"))
		require.NoError(t, err)
		return signed
	}
	send := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/catalog", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		s.Router.ServeHTTP(rec, req)
		return rec
	}

	assert.Equal(t, "MISS", send(token("alice", "read")).Header().Get("X-Cache"))
	assert.Equal(t, "HIT", send(token("bob", "read")).Header().Get("X-Cache"), "another allowed caller shares it")
	assert.Equal(t, http.StatusUnauthorized, send("").Code)
	assert.Equal(t, http.StatusUnauthorized, send("not-a-token").Code)
	assert.Equal(t, http.StatusForbidden, send(token("eve", "write")).Code)
	assert.Equal(t, 1, calls)
}

// Tokens from an identity provider are verified with the keys it publishes,
// wired the way main wires them.
func TestInitializeServerVerifiesTokensWithTheJWKS(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	b64 := func(n *big.Int) string { return base64.RawURLEncoding.EncodeToString(n.FillBytes(make([]byte, 32))) }
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"keys": [{"kty": "EC", "kid": "k1", "crv": "P-256", "x": %q, "y": %q}]}`, b64(key.X), b64(key.Y))
	}))
	t.Cleanup(provider.Close)

	cfg, err := config.Parse([]byte(`
Plugins: {JWTConfig: {JWKSURL: "` + provider.URL + `/jwks.json", Audience: gateway}}
Services:
  api:
    URL: http://upstream.test
    Endpoints: [{HTTPMethod: GET, Path: /me, Auth: {ApplyAuth: true}}]
`))
	require.NoError(t, err)
	s, err := InitializeServer(ServerDependencies{
		loadConfig: loads(cfg, nil),
		router:     mux.NewRouter(),
		newProxy: func(string, time.Duration) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), nil
		},
	}, "config.yaml")
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, closer := range s.Closers {
			closer.Close()
		}
	})

	claims := auth.Claims{RegisteredClaims: jwt.RegisteredClaims{Audience: []string{"gateway"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
	signed := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signed.Header["kid"] = "k1"
	fromProvider, err := signed.SignedString(key)
	require.NoError(t, err)
	withSecret, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("any-secret"))
	require.NoError(t, err)

	for token, status := range map[string]int{fromProvider: http.StatusOK, withSecret: http.StatusUnauthorized} {
		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		s.Router.ServeHTTP(rec, req)
		assert.Equal(t, status, rec.Code)
	}
}

// main's own wiring must start the gateway from the shipped config.
func TestDefaultDependenciesStartTheShippedConfig(t *testing.T) {
	t.Setenv("CONFIG_PATH", "config.yaml")

	s, err := InitializeServer(defaultDependencies(), resolveConfigPath(""))
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, closer := range s.Closers {
			closer.Close()
		}
	})

	assert.Equal(t, ":4000", s.Addr)
	assert.NotEmpty(t, s.Closers, "the shipped config uses the logger")
}

func TestInitializeServerAnswersCorsPreflights(t *testing.T) {
	cfg, err := config.Parse([]byte(`
Plugins: {CORSConfig: {Origins: "https://app.test", Methods: "GET, POST", Headers: "Content-Type"}}
Services:
  api:
    URL: http://upstream.test
    Plugins: [cors]
    Endpoints: [{HTTPMethod: POST, Path: /items}]
`))
	require.NoError(t, err)
	s, err := InitializeServer(ServerDependencies{
		loadConfig: loads(cfg, nil),
		router:     mux.NewRouter(),
		newProxy: func(string, time.Duration) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("a preflight must not reach the upstream") }), nil
		},
	}, "config.yaml")
	require.NoError(t, err)

	req := httptest.NewRequest("OPTIONS", "/items", nil)
	req.Header.Set("Origin", "https://app.test")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://app.test", rec.Header().Get("Access-Control-Allow-Origin"))
}

// Every response the gateway answers itself, rather than the upstream, is
// JSON with a message, whichever part of the gateway rejected the request.
func TestInitializeServerAnswersErrorsAsJSON(t *testing.T) {
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()
	cfg, err := config.Parse([]byte(`
Server: {CircuitBreaker: {ConsecutiveFailures: 1}}
Plugins:
  CORSConfig: {Origins: "https://app.test", Methods: "GET"}
  RateLimitConfig: {Limit: 1, Window: 1m}
  JWTConfig: {SecretKey: test-secret}
Services:
  api:
    URL: ` + down.URL + `
    Plugins: [cors, rate-limit]
    Endpoints:
      - {HTTPMethod: GET, Path: /items}
      - {HTTPMethod: GET, Path: /secure, Auth: {ApplyAuth: true}}
`))
	require.NoError(t, err)
	s, err := InitializeServer(ServerDependencies{
		loadConfig: loads(cfg, nil),
		router:     mux.NewRouter(),
		newProxy:   proxy.NewHandler,
		newStore:   func(config.StoreConfig) (store.Store, error) { return memory.New(), nil },
	}, "config.yaml")
	require.NoError(t, err)

	send := func(method, path, client, origin string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.RemoteAddr = client + ":1234"
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		s.Router.ServeHTTP(rec, req)
		return rec
	}

	tests := []struct {
		name    string
		rec     *httptest.ResponseRecorder
		status  int
		message string
	}{
		{"unknown path", send("GET", "/missing", "10.0.0.1", ""), http.StatusNotFound, "Not found"},
		{"unknown method", send("DELETE", "/items", "10.0.0.2", ""), http.StatusMethodNotAllowed, "Method not allowed"},
		{"origin not allowed", send("GET", "/items", "10.0.0.3", "https://evil.test"), http.StatusForbidden, "CORS policy violation: Invalid Origin"},
		{"no token", send("GET", "/secure", "10.0.0.4", ""), http.StatusUnauthorized, "Authorization header is required"},
		{"upstream down", send("GET", "/items", "10.0.0.5", ""), http.StatusBadGateway, "Bad gateway"},
		{"rate limited", send("GET", "/items", "10.0.0.5", ""), http.StatusTooManyRequests, "Rate limit exceeded"},
		{"breaker open", send("GET", "/items", "10.0.0.6", ""), http.StatusServiceUnavailable, "Service unavailable"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.status, tc.rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", tc.rec.Header().Get("Content-Type"))
			assert.JSONEq(t, `{"message": "`+tc.message+`"}`, tc.rec.Body.String())
		})
	}
}
