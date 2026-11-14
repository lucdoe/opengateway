// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validServer() config.ServerConfig {
	return config.ServerConfig{
		Port:              4000,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
		ShutdownTimeout:   15 * time.Second,
		UpstreamTimeout:   30 * time.Second,
		CircuitBreaker:    config.CircuitBreakerConfig{ConsecutiveFailures: 5, OpenTimeout: 30 * time.Second, HalfOpenRequests: 1},
	}
}

// validConfig returns a configuration that uses every plugin and passes
// validation. Tests break one thing at a time.
func validConfig() *config.Config {
	return &config.Config{
		Server: validServer(),
		Store:  config.StoreConfig{Type: "redis", Redis: config.RedisConfig{Host: "localhost", Port: 6379}},
		Plugins: config.PluginConfig{
			LoggerConfig:    config.LoggerConfig{FilePath: "server.log"},
			JWTConfig:       config.JWTConfig{SecretKey: "secret", SigningMethod: "HS256"},
			RateLimitConfig: config.RateLimitConfig{Limit: 100, Window: time.Minute},
			CORSConfig:      config.CORSConfig{Origins: "*", Methods: "GET, POST"},
		},
		Services: map[string]config.Service{
			"api": {
				URL:     "https://example.com",
				Subpath: "/api",
				Plugins: []string{"logger", "cors", "rate-limit", "cache"},
				Endpoints: []config.Endpoint{
					{Name: "List", HTTPMethod: "GET", Path: "/items"},
					{Name: "Create", HTTPMethod: "POST", Path: "/items", Auth: config.AuthConfig{ApplyAuth: true}},
				},
			},
		},
	}
}

func TestValidateAcceptsAValidConfig(t *testing.T) {
	assert.NoError(t, validConfig().Validate())
}

func TestValidateServicesAndEndpoints(t *testing.T) {
	// edit changes the "api" service of a valid config.
	tests := []struct {
		name     string
		edit     func(s *config.Service)
		expected string
	}{
		{"missing URL", func(s *config.Service) { s.URL = "" },
			`service "api": URL is required`},
		{"URL without scheme", func(s *config.Service) { s.URL = "example.com" },
			`service "api": URL "example.com" must be an absolute http or https URL`},
		{"URL with unsupported scheme", func(s *config.Service) { s.URL = "ftp://example.com" },
			`must be an absolute http or https URL`},
		{"URL that does not parse", func(s *config.Service) { s.URL = "http://exa mple.com" },
			`must be an absolute http or https URL`},
		{"Subpath without leading slash", func(s *config.Service) { s.Subpath = "api" },
			`service "api": Subpath "api" must start with "/"`},
		{"Subpath ending in a slash", func(s *config.Service) { s.Subpath = "/api/" },
			`service "api": Subpath "/api/" must not end with "/"`},
		{"no endpoints", func(s *config.Service) { s.Endpoints = nil },
			`service "api": at least one endpoint is required`},
		{"unknown service plugin", func(s *config.Service) { s.Plugins = []string{"loger"} },
			`service "api": unknown plugin "loger"`},
		{"missing method", func(s *config.Service) { s.Endpoints[0].HTTPMethod = "" },
			`service "api", endpoint "List": HTTPMethod is required`},
		{"unknown method", func(s *config.Service) { s.Endpoints[0].HTTPMethod = "FETCH" },
			`service "api", endpoint "List": HTTPMethod "FETCH" is not one of GET`},
		{"missing path", func(s *config.Service) { s.Endpoints[0].Path = "" },
			`service "api", endpoint "List": Path is required`},
		{"path without leading slash", func(s *config.Service) { s.Endpoints[0].Path = "items" },
			`service "api", endpoint "List": Path "items" must start with "/"`},
		{"reserved path", func(s *config.Service) { s.Subpath = ""; s.Endpoints[0].Path = "/__health" },
			`service "api", endpoint "List": path /__health is reserved: paths starting with /__ belong to the gateway`},
		{"reserved path through the Subpath", func(s *config.Service) { s.Subpath = "/__internal" },
			`service "api", endpoint "List": path /__internal/items is reserved`},
		{"unknown endpoint plugin", func(s *config.Service) { s.Endpoints[1].Plugins = []string{"cash"} },
			`service "api", endpoint "Create": unknown plugin "cash"`},
		{"unnamed endpoint is identified by position", func(s *config.Service) { s.Endpoints[1].Name = ""; s.Endpoints[1].Path = "" },
			`service "api", endpoint #2: Path is required`},
		{"duplicate route in one service", func(s *config.Service) {
			s.Endpoints = append(s.Endpoints, config.Endpoint{Name: "Again", HTTPMethod: "get", Path: "/items"})
		}, `service "api", endpoint "Again": route GET /api/items is already used by service "api", endpoint "List"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			service := cfg.Services["api"]
			tc.edit(&service)
			cfg.Services["api"] = service

			err := cfg.Validate()

			require.Error(t, err)
			assert.ErrorContains(t, err, "invalid configuration")
			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestValidateAcceptsLowerCaseMethods(t *testing.T) {
	cfg := validConfig()
	cfg.Services["api"].Endpoints[0].HTTPMethod = "get"

	assert.NoError(t, cfg.Validate())
}

func TestValidateRejectsTheSameRouteInTwoServices(t *testing.T) {
	cfg := validConfig()
	cfg.Services["other"] = config.Service{
		URL:     "http://localhost:3001",
		Subpath: "/api",
		Endpoints: []config.Endpoint{
			{Name: "Other List", HTTPMethod: "GET", Path: "/items"},
			{Name: "Other Delete", HTTPMethod: "DELETE", Path: "/items"},
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.ErrorContains(t, err, `service "other", endpoint "Other List": route GET /api/items is already used by service "api", endpoint "List"`)
	assert.NotContains(t, err.Error(), "Other Delete")
}

func TestValidateServerAndRedisSettings(t *testing.T) {
	tests := []struct {
		name     string
		edit     func(c *config.Config)
		expected string
	}{
		{"port of zero", func(c *config.Config) { c.Server.Port = 0 }, "Server.Port 0 must be between 1 and 65535"},
		{"port too large", func(c *config.Config) { c.Server.Port = 70000 }, "Server.Port 70000 must be between 1 and 65535"},
		{"negative port", func(c *config.Config) { c.Server.Port = -1 }, "Server.Port -1 must be between 1 and 65535"},
		{"no read header timeout", func(c *config.Config) { c.Server.ReadHeaderTimeout = 0 }, "Server.ReadHeaderTimeout 0s must be greater than 0"},
		{"negative idle timeout", func(c *config.Config) { c.Server.IdleTimeout = -time.Second }, "Server.IdleTimeout -1s must be greater than 0"},
		{"no shutdown timeout", func(c *config.Config) { c.Server.ShutdownTimeout = 0 }, "Server.ShutdownTimeout 0s must be greater than 0"},
		{"negative upstream timeout", func(c *config.Config) { c.Server.UpstreamTimeout = -time.Minute }, "Server.UpstreamTimeout -1m0s must be greater than 0"},
		{"no breaker failures", func(c *config.Config) { c.Server.CircuitBreaker.ConsecutiveFailures = 0 }, "Server.CircuitBreaker.ConsecutiveFailures 0 must be greater than 0"},
		{"negative breaker open timeout", func(c *config.Config) { c.Server.CircuitBreaker.OpenTimeout = -time.Second }, "Server.CircuitBreaker.OpenTimeout -1s must be greater than 0"},
		{"negative breaker half-open requests", func(c *config.Config) { c.Server.CircuitBreaker.HalfOpenRequests = -1 }, "Server.CircuitBreaker.HalfOpenRequests -1 must be greater than 0"},
		{"redis without host", func(c *config.Config) { c.Store.Redis.Host = "" }, "Store.Redis.Host is required because Store.Type is redis"},
		{"redis port too large", func(c *config.Config) { c.Store.Redis.Port = 70000 }, "Store.Redis.Port 70000 must be between 1 and 65535"},
		{"unknown store", func(c *config.Config) { c.Store.Type = "memcached" }, `Store.Type "memcached" is not one of memory, redis`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			tc.edit(cfg)

			assert.ErrorContains(t, cfg.Validate(), tc.expected)
		})
	}
}

func TestValidateIgnoresRedisSettingsWhenNoPluginNeedsRedis(t *testing.T) {
	cfg := validConfig()
	service := cfg.Services["api"]
	service.Plugins = []string{"logger", "cors"}
	cfg.Services["api"] = service
	cfg.Store.Redis = config.RedisConfig{}

	assert.NoError(t, cfg.Validate())
}

func TestValidateAcceptsTheLoggerWithoutAFile(t *testing.T) {
	cfg := validConfig()
	cfg.Plugins.LoggerConfig.FilePath = "" // logs to stdout

	assert.NoError(t, cfg.Validate())
}

func TestValidateIgnoresRedisSettingsOfTheMemoryStore(t *testing.T) {
	cfg := validConfig()
	cfg.Store = config.StoreConfig{Type: "memory"}

	assert.NoError(t, cfg.Validate())
}

func TestValidateEndpointScope(t *testing.T) {
	withScope := func(service config.Service, auth config.AuthConfig, plugins ...string) *config.Config {
		cfg := validConfig()
		service.Endpoints = []config.Endpoint{{Name: "Admin", HTTPMethod: "GET", Path: "/admin", Auth: auth, Plugins: plugins}}
		cfg.Services["api"] = service
		return cfg
	}
	service := validConfig().Services["api"]
	serviceWithAuth := service
	serviceWithAuth.Plugins = append([]string{"auth"}, service.Plugins...)

	assert.NoError(t, withScope(service, config.AuthConfig{ApplyAuth: true, Scope: "admin"}).Validate(), "ApplyAuth")
	assert.NoError(t, withScope(service, config.AuthConfig{Scope: "admin"}, "auth").Validate(), "endpoint plugin")
	assert.NoError(t, withScope(serviceWithAuth, config.AuthConfig{Scope: "admin"}).Validate(), "service plugin")
	assert.ErrorContains(t, withScope(service, config.AuthConfig{Scope: "admin"}).Validate(),
		`service "api", endpoint "Admin": Auth.Scope needs auth: set Auth.ApplyAuth or add the auth plugin`)
}

func TestValidateRejectsAnEmptyConfig(t *testing.T) {
	assert.ErrorContains(t, (&config.Config{}).Validate(), "no services are configured")
}

func TestValidatePluginSettings(t *testing.T) {
	tests := []struct {
		name     string
		edit     func(p *config.PluginConfig)
		expected string
	}{
		{"cors without origins", func(p *config.PluginConfig) { p.CORSConfig.Origins = "" },
			"Plugins.CORSConfig.Origins is required"},
		{"cors without methods", func(p *config.PluginConfig) { p.CORSConfig.Methods = "" },
			"Plugins.CORSConfig.Methods is required"},
		{"rate limit of zero", func(p *config.PluginConfig) { p.RateLimitConfig.Limit = 0 },
			"Plugins.RateLimitConfig.Limit must be greater than 0"},
		{"negative rate limit window", func(p *config.PluginConfig) { p.RateLimitConfig.Window = -time.Second },
			"Plugins.RateLimitConfig.Window must be greater than 0"},
		{"unknown rate-limit key", func(p *config.PluginConfig) { p.RateLimitConfig.Key = "user" },
			`Plugins.RateLimitConfig.Key "user" is not one of ip, header:Name-Of-Header or jwt:sub`},
		{"rate-limit key with an empty header name", func(p *config.PluginConfig) { p.RateLimitConfig.Key = "header:" },
			`Plugins.RateLimitConfig.Key "header:" is not one of`},
		{"rate-limit key with an invalid header name", func(p *config.PluginConfig) { p.RateLimitConfig.Key = "header:X Api Key" },
			`Plugins.RateLimitConfig.Key "header:X Api Key" is not one of`},
		{"auth without secret or JWKS", func(p *config.PluginConfig) { p.JWTConfig.SecretKey = "" },
			"Plugins.JWTConfig needs SecretKey (for HS256 tokens) or JWKSURL"},
		{"both secret and JWKS", func(p *config.PluginConfig) { p.JWTConfig.JWKSURL = "https://auth.example.com/jwks.json" },
			"Plugins.JWTConfig sets both SecretKey and JWKSURL; set one (JWT_SECRET also sets SecretKey)"},
		{"key algorithm with a secret", func(p *config.PluginConfig) { p.JWTConfig.SigningMethod = "RS256" },
			`Plugins.JWTConfig.SigningMethod "RS256" is not one of HS256, HS384, HS512`},
		{"HMAC algorithm with JWKS", func(p *config.PluginConfig) {
			p.JWTConfig.SecretKey, p.JWTConfig.JWKSURL, p.JWTConfig.SigningMethod = "", "https://auth.example.com/jwks.json", "HS256"
		}, `Plugins.JWTConfig.SigningMethod "HS256" is not one of RS256, RS384, RS512, PS256, PS384, PS512, ES256, ES384, ES512, EdDSA`},
		{"JWKS that is not a URL", func(p *config.PluginConfig) { p.JWTConfig.SecretKey, p.JWTConfig.JWKSURL = "", "auth.example.com/jwks" },
			`Plugins.JWTConfig.JWKSURL "auth.example.com/jwks" must be an absolute https URL`},
		{"JWKS over plain http to an outside host", func(p *config.PluginConfig) {
			p.JWTConfig.SecretKey, p.JWTConfig.JWKSURL = "", "http://auth.example.com/jwks.json"
		}, `Plugins.JWTConfig.JWKSURL "http://auth.example.com/jwks.json" must use https: keys fetched over plain http can be swapped to forge tokens`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			tc.edit(&cfg.Plugins)

			assert.ErrorContains(t, cfg.Validate(), tc.expected)
		})
	}
}

func TestValidateAcceptsJWKSSettings(t *testing.T) {
	for _, jwt := range []config.JWTConfig{
		{JWKSURL: "https://auth.example.com/.well-known/jwks.json"},
		{JWKSURL: "https://auth.example.com/jwks.json", SigningMethod: "ES256"},
		{JWKSURL: "http://keycloak:8080/realms/app/protocol/openid-connect/certs"}, // plain http inside, a Compose service
		{SecretKey: "secret", SigningMethod: "HS512"},
	} {
		cfg := validConfig()
		cfg.Plugins.JWTConfig = jwt
		assert.NoError(t, cfg.Validate(), "%+v", jwt)
	}
}

func TestValidateAcceptsRateLimitKeys(t *testing.T) {
	for _, key := range []string{"", "ip", "header:CF-Connecting-IP", "header:X-API-Key", "jwt:sub"} {
		cfg := validConfig()
		cfg.Plugins.RateLimitConfig.Key = key
		assert.NoError(t, cfg.Validate(), key)
	}
}

func TestPluginsInOrderRunsRateLimitAfterAuthOnlyForTokenSubjects(t *testing.T) {
	cfg := validConfig()
	assert.Equal(t, config.PluginOrder, cfg.PluginsInOrder())

	cfg.Plugins.RateLimitConfig.Key = "header:X-API-Key"
	assert.Equal(t, config.PluginOrder, cfg.PluginsInOrder())

	cfg.Plugins.RateLimitConfig.Key = "jwt:sub"
	assert.Equal(t, []string{"logger", "cors", "auth", "rate-limit", "cache"}, cfg.PluginsInOrder())
}

func TestJWTAlgorithms(t *testing.T) {
	assert.Equal(t, []string{"HS256"}, config.JWTConfig{SecretKey: "s"}.Algorithms())
	assert.Equal(t, []string{"HS512"}, config.JWTConfig{SecretKey: "s", SigningMethod: "HS512"}.Algorithms())
	assert.Equal(t, config.KeyAlgorithms, config.JWTConfig{JWKSURL: "https://a.test/jwks"}.Algorithms())
	assert.Equal(t, []string{"ES256"}, config.JWTConfig{JWKSURL: "https://a.test/jwks", SigningMethod: "ES256"}.Algorithms())
}

func TestValidateIgnoresSettingsOfUnusedPlugins(t *testing.T) {
	cfg := &config.Config{
		Server: validServer(),
		Store:  config.StoreConfig{Type: "memory"},
		Services: map[string]config.Service{
			"api": {
				URL:       "https://example.com",
				Plugins:   []string{"cache"},
				Endpoints: []config.Endpoint{{Name: "List", HTTPMethod: "GET", Path: "/items"}},
			},
		},
	}

	assert.NoError(t, cfg.Validate())
}

func TestValidateReportsEveryProblemAtOnce(t *testing.T) {
	cfg := validConfig()
	cfg.Plugins.RateLimitConfig.Limit = 0
	service := cfg.Services["api"]
	service.URL = ""
	service.Endpoints[0].Path = "items"
	cfg.Services["api"] = service

	err := cfg.Validate()

	require.Error(t, err)
	assert.Equal(t, 3, strings.Count(err.Error(), "\n- "))
	assert.ErrorContains(t, err, "URL is required")
	assert.ErrorContains(t, err, `Path "items" must start with "/"`)
	assert.ErrorContains(t, err, "Limit must be greater than 0")
}

func TestUsesPlugin(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"a": {Plugins: []string{"logger"}, Endpoints: []config.Endpoint{
				{Name: "One", Plugins: []string{"cache"}},
				{Name: "Two", Auth: config.AuthConfig{ApplyAuth: true}},
			}},
		},
	}

	assert.True(t, cfg.UsesPlugin(config.PluginLogger), "listed on a service")
	assert.True(t, cfg.UsesPlugin(config.PluginCache), "listed on an endpoint")
	assert.True(t, cfg.UsesPlugin(config.PluginAuth), "enabled through ApplyAuth")
	assert.False(t, cfg.UsesPlugin(config.PluginCORS))
	assert.False(t, cfg.UsesPlugin(config.PluginRateLimit))
}
