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

const layeredConfig = `
Server: {UpstreamTimeout: 30s}
Plugins:
  RateLimit: {Limit: 100, Window: 1m}
  CORS: {Origins: "*", Methods: "GET"}
Services:
  accounts:
    URL: http://accounts
    Subpath: /accounts
    Plugins: [rate-limit, cors]
    Settings:
      UpstreamTimeout: 5s
      RateLimit: {Limit: 10}
    Endpoints:
      - {Name: Profile, HTTPMethod: GET, Path: /me}
      - Name: Login
        HTTPMethod: post
        Path: /login
        Settings:
          RateLimit: {Window: 1h, Key: "header:X-Forwarded-User"}
          CORS: {Origins: "https://app.example.com", Methods: "POST"}
  news:
    URL: http://news
    Plugins: [rate-limit]
    Endpoints: [{HTTPMethod: GET, Path: /news}]
`

func layered(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := config.Parse([]byte(layeredConfig))
	require.NoError(t, err)
	return cfg
}

func TestRouteSettingsWithoutSettingsAreTheGlobalOnes(t *testing.T) {
	cfg := layered(t)

	route := cfg.RouteSettings("news", cfg.Services["news"].Endpoints[0])

	assert.Equal(t, config.RouteSettings{
		RateLimit: config.RateLimitConfig{Limit: 100, Window: time.Minute},
		CORS:      config.CORSConfig{Origins: "*", Methods: "GET"},
	}, route)
	assert.Equal(t, 30*time.Second, cfg.UpstreamTimeout("news"))
}

func TestRouteSettingsMergeTheServiceOverTheGlobalOnes(t *testing.T) {
	cfg := layered(t)

	route := cfg.RouteSettings("accounts", cfg.Services["accounts"].Endpoints[0])

	assert.Equal(t, config.RateLimitConfig{Limit: 10, Window: time.Minute}, route.RateLimit, "only Limit is set, Window stays")
	assert.Equal(t, "accounts", route.RateLimitScope)
	assert.Equal(t, config.CORSConfig{Origins: "*", Methods: "GET"}, route.CORS)
	assert.Empty(t, route.CORSScope, "the service changes nothing about cors")
	assert.Equal(t, 5*time.Second, cfg.UpstreamTimeout("accounts"))
}

func TestRouteSettingsMergeTheEndpointOverItsService(t *testing.T) {
	cfg := layered(t)

	route := cfg.RouteSettings("accounts", cfg.Services["accounts"].Endpoints[1])

	assert.Equal(t, config.RateLimitConfig{Limit: 10, Window: time.Hour, Key: "header:X-Forwarded-User"}, route.RateLimit, "Limit from the service, the rest from the endpoint")
	assert.Equal(t, "accounts POST /accounts/login", route.RateLimitScope)
	assert.Equal(t, config.CORSConfig{Origins: "https://app.example.com", Methods: "POST"}, route.CORS)
	assert.Equal(t, "accounts POST /accounts/login", route.CORSScope)
}

func TestValidateChecksTheSettingsEveryRouteRunsWith(t *testing.T) {
	cfg, err := config.Parse([]byte(`
Plugins:
  CORS: {Origins: "*"}
Services:
  accounts:
    URL: http://accounts
    Plugins: [rate-limit]
    Settings:
      UpstreamTimeout: -1s
      RateLimit: {Limit: 5, Key: user}
    Endpoints:
      - Name: Login
        HTTPMethod: POST
        Path: /login
        Plugins: [cors]
        Settings:
          UpstreamTimeout: 2s
          CORS: {Headers: "Content-Type"}
      - {Name: Logout, HTTPMethod: POST, Path: /logout}
  news:
    URL: http://news
    Plugins: [rate-limit]
    Endpoints: [{HTTPMethod: GET, Path: /news}]
`))

	assert.Nil(t, cfg)
	for _, want := range []string{
		`service "accounts": Settings.UpstreamTimeout -1s must be greater than 0`,
		`service "accounts", endpoint "Login": Settings.UpstreamTimeout can only be set for the service`,
		`service "accounts", endpoint "Login": RateLimit.Window must be greater than 0 because the rate-limit plugin is used`,
		`service "accounts", endpoint "Login": RateLimit.Key "user" is not one of`,
		`service "accounts", endpoint "Login": CORS.Methods is required because the cors plugin is used`,
		`service "accounts", endpoint "Logout": RateLimit.Window must be greater than 0`,
		`Plugins.RateLimit.Limit must be greater than 0 because the rate-limit plugin is used`, // for news
	} {
		assert.ErrorContains(t, err, want)
	}
}

func TestValidateAcceptsGlobalSettingsLeftOutWhereEveryRouteSetsThem(t *testing.T) {
	_, err := config.Parse([]byte(`
Services:
  accounts:
    URL: http://accounts
    Plugins: [rate-limit]
    Settings:
      RateLimit: {Limit: 5, Window: 1m}
    Endpoints: [{HTTPMethod: POST, Path: /login}]
`))

	assert.NoError(t, err)
}

func TestValidateReportsAProblemOfTheGlobalSettingsOnce(t *testing.T) {
	_, err := config.Parse([]byte(`
Services:
  api:
    URL: http://api
    Plugins: [rate-limit]
    Endpoints:
      - {HTTPMethod: GET, Path: /a}
      - {HTTPMethod: GET, Path: /b}
`))

	require.Error(t, err)
	assert.Equal(t, 1, strings.Count(err.Error(), "Plugins.RateLimit.Limit must be greater than 0"))
}
