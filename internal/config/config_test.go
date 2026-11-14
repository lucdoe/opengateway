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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fullConfig = `
Plugins:
  LoggerConfig:
    FilePath: "server.log"
    ErrOutput: "stderr"
  JWTConfig:
    SecretKey: "secret"
    SigningMethod: "HS256"
    Issuer: "issuer"
    Audience: "audience"
    Scope: "scope"
  RateLimitConfig:
    Limit: 100
    Window: 1m
  CORSConfig:
    Origins: "*"
    Methods: "GET, POST"
    Headers: "Content-Type"

Services:
  service1:
    URL: "http://example.org"
    Protocol: "HTTPS"
    Subpath: "/api"
    Plugins:
      - "logger"
      - "cors"
      - "rate-limit"
      - "cache"
    Endpoints:
      - Name: "Submit"
        HTTPMethod: "POST"
        Path: "/submit"
        QueryParams:
          - Key: "key"
            Value: "value"
        Auth:
          ApplyAuth: true
          Method: "Bearer"
          Algorithm: "HS256"
          Scope: "read:write"
          SecretKey: "secret"
        Body:
          key1: value1
        Plugins:
          - "logger"
`

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoadReadsEveryField(t *testing.T) {
	cfg, err := config.Load(writeConfig(t, fullConfig))
	require.NoError(t, err)

	assert.Equal(t, config.PluginConfig{
		LoggerConfig:    config.LoggerConfig{FilePath: "server.log", ErrOutput: "stderr"},
		JWTConfig:       config.JWTConfig{SecretKey: "secret", SigningMethod: "HS256", Issuer: "issuer", Audience: "audience", Scope: "scope"},
		RateLimitConfig: config.RateLimitConfig{Limit: 100, Window: time.Minute},
		CORSConfig:      config.CORSConfig{Origins: "*", Methods: "GET, POST", Headers: "Content-Type"},
	}, cfg.Plugins)

	require.Len(t, cfg.Services, 1)
	assert.Equal(t, config.Service{
		URL:      "http://example.org",
		Protocol: "HTTPS",
		Subpath:  "/api",
		Plugins:  []string{"logger", "cors", "rate-limit", "cache"},
		Endpoints: []config.Endpoint{{
			Name:        "Submit",
			HTTPMethod:  "POST",
			Path:        "/submit",
			QueryParams: []config.QueryParam{{Key: "key", Value: "value"}},
			Auth:        config.AuthConfig{ApplyAuth: true, Method: "Bearer", Algorithm: "HS256", Scope: "read:write", SecretKey: "secret"},
			Body:        map[string]interface{}{"key1": "value1"},
			Plugins:     []string{"logger"},
		}},
	}, cfg.Services["service1"])
}

// The config file that ships with the gateway must always load.
func TestLoadAcceptsTheShippedConfig(t *testing.T) {
	cfg, err := config.Load("../../cmd/opengateway/config.yaml")

	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Services)
}

func TestLoadFailsWhenTheFileIsMissing(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))

	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.ErrorContains(t, err, "reading config")
}

func TestLoadNamesTheFileInErrors(t *testing.T) {
	path := writeConfig(t, "Services: {}")

	_, err := config.Load(path)

	assert.ErrorContains(t, err, path)
	assert.ErrorContains(t, err, "no services are configured")
}

func TestParseRejectsMalformedAndUnknownInput(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name:     "services as a list",
			yaml:     "Services:\n  - URL: http://example.com\n",
			expected: "cannot unmarshal",
		},
		{
			name:     "not YAML",
			yaml:     "Services: [unclosed",
			expected: "yaml",
		},
		{
			name:     "misspelled top-level key",
			yaml:     "Service:\n  a:\n    URL: http://example.com\n",
			expected: "field Service not found",
		},
		{
			name: "misspelled endpoint key",
			yaml: `
Services:
  a:
    URL: http://example.com
    Endpoints:
      - Name: Items
        HTTPMethod: GET
        Path: /items
        Plugin: ["cache"]
`,
			expected: "field Plugin not found",
		},
		{
			name:     "empty file",
			yaml:     "",
			expected: "no services are configured",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Parse([]byte(tc.yaml))

			assert.Nil(t, cfg)
			assert.ErrorContains(t, err, tc.expected)
		})
	}
}
