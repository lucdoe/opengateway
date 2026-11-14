// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"strings"
	"testing"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// strongSecret is long enough for HS256.
var strongSecret = strings.Repeat("k", 32)

// authConfig is a config with one auth endpoint on url, with the given JWT
// settings.
func authConfig(t *testing.T, url string, jwt config.JWTConfig) *config.Config {
	t.Helper()
	return &config.Config{
		Plugins: config.PluginConfig{JWTConfig: jwt},
		Services: map[string]config.Service{
			"users": {URL: url, Endpoints: []config.Endpoint{{HTTPMethod: "GET", Path: "/me", Plugins: []string{"auth"}}}},
		},
	}
}

func TestWarningsAreEmptyForASafeAuthConfig(t *testing.T) {
	cfg := authConfig(t, "https://users.example.com", config.JWTConfig{SecretKey: strongSecret, Issuer: "auth", Audience: "gateway"})

	assert.Empty(t, cfg.Warnings())
}

func TestWarningsAreEmptyWithoutAuth(t *testing.T) {
	cfg, err := config.Parse([]byte(`
Plugins: {JWT: {SecretKey: your-256-bit-secret}}
Services:
  api:
    URL: http://api.example.com
    Endpoints: [{HTTPMethod: GET, Path: /items}]
`))
	require.NoError(t, err)

	assert.Empty(t, cfg.Warnings(), "the JWT settings and plain http only matter where tokens are used")
}

func TestWarningsFlagWeakOrExampleSecrets(t *testing.T) {
	example := authConfig(t, "https://users.example.com", config.JWTConfig{SecretKey: "your-256-bit-secret", Issuer: "auth", Audience: "gateway"})
	short := authConfig(t, "https://users.example.com", config.JWTConfig{SecretKey: "secret", Issuer: "auth", Audience: "gateway"})

	assert.Equal(t, []string{"Plugins.JWT.SecretKey is the example secret from the docs, so anyone can sign tokens with it; set your own, for example with JWT_SECRET"}, example.Warnings())
	assert.Equal(t, []string{"Plugins.JWT.SecretKey is 6 bytes; HS256 needs at least 32 (256 bits) to resist guessing"}, short.Warnings())
}

func TestWarningsFlagMissingIssuerAndAudience(t *testing.T) {
	cfg := authConfig(t, "https://users.example.com", config.JWTConfig{SecretKey: strongSecret})

	assert.Equal(t, []string{
		"Plugins.JWT.Issuer is empty, so tokens from any issuer that has the secret are accepted",
		"Plugins.JWT.Audience is empty, so tokens meant for other services that share the secret are accepted",
	}, cfg.Warnings())
}

func TestWarningsWithJWKSSkipTheSecretAndNameTheIdentityProvider(t *testing.T) {
	safe := authConfig(t, "https://users.example.com", config.JWTConfig{JWKSURL: "https://auth.example.com/jwks.json", Issuer: "https://auth.example.com/", Audience: "gateway"})
	open := authConfig(t, "https://users.example.com", config.JWTConfig{JWKSURL: "https://auth.example.com/jwks.json"})

	assert.Empty(t, safe.Warnings(), "no secret is used, so its length does not matter")
	assert.Equal(t, []string{
		"Plugins.JWT.Issuer is empty, so tokens from any issuer whose keys are in the key set are accepted",
		"Plugins.JWT.Audience is empty, so tokens meant for other apps of the same identity provider are accepted",
	}, open.Warnings())
}

func TestWarningsFlagTokensSentOverPlainHTTPToOutsideHosts(t *testing.T) {
	safe := config.JWTConfig{SecretKey: strongSecret, Issuer: "auth", Audience: "gateway"}

	for _, url := range []string{"http://users.example.com", "http://203.0.113.5:8080", "http://Users.Example.COM./api"} {
		assert.Len(t, authConfig(t, url, safe).Warnings(), 1, url)
	}
	assert.Equal(t, []string{`service "users": tokens are forwarded over plain http to users.example.com; use https`},
		authConfig(t, "http://users.example.com", safe).Warnings())

	for _, url := range []string{
		"https://users.example.com",
		"http://localhost:3001", "http://127.0.0.1:3001", "http://[::1]:3001",
		"http://10.0.0.7", "http://192.168.1.20", "http://172.16.0.3",
		"http://users", // a Compose or Kubernetes service name
		"http://users.internal", "http://printer.local", "http://users.default.svc", "http://users.default.svc.cluster.local", "http://app.localhost",
	} {
		assert.Empty(t, authConfig(t, url, safe).Warnings(), url)
	}
}

func TestWarningsOnlyFlagPlainHTTPForServicesThatUseAuth(t *testing.T) {
	cfg := authConfig(t, "https://users.example.com", config.JWTConfig{SecretKey: strongSecret, Issuer: "auth", Audience: "gateway"})
	cfg.Services["public"] = config.Service{URL: "http://public.example.com", Endpoints: []config.Endpoint{{HTTPMethod: "GET", Path: "/news"}}}

	assert.Empty(t, cfg.Warnings())
}
