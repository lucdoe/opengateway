// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withService is a minimal valid config with extra put at the top level.
func withService(extra string) []byte {
	return []byte(extra + `
Services:
  api:
    URL: "${API_URL:-http://localhost:3001}"
    Endpoints: [{HTTPMethod: GET, Path: /items}]
`)
}

func TestParseFillsPlaceholdersFromTheEnvironment(t *testing.T) {
	t.Setenv("API_URL", "https://api.internal")
	t.Setenv("GATEWAY_PORT", "4001")
	t.Setenv("WINDOW", "30s")
	t.Setenv("SECRET", "s3cr#t: \"quoted\"")
	t.Setenv("REGION", "eu")

	cfg, err := config.Parse(withService(`
Server:
  Port: ${GATEWAY_PORT}
Plugins:
  RateLimit:
    Limit: 10
    Window: ${WINDOW}
  JWT:
    SecretKey: "${SECRET}"
    Issuer: 'https://${REGION}.auth.test/'
`))
	require.NoError(t, err)

	assert.Equal(t, "https://api.internal", cfg.Services["api"].URL)
	assert.Equal(t, 4001, cfg.Server.Port, "an unquoted placeholder is typed like the value it becomes")
	assert.Equal(t, 30*time.Second, cfg.Plugins.RateLimitConfig.Window)
	assert.Equal(t, `s3cr#t: "quoted"`, cfg.Plugins.JWTConfig.SecretKey, "a value is taken as it is, never parsed as YAML")
	assert.Equal(t, "https://eu.auth.test/", cfg.Plugins.JWTConfig.Issuer)
}

func TestParseUsesTheFallbackWhenTheVariableIsUnsetOrEmpty(t *testing.T) {
	t.Setenv("API_URL", "")

	cfg, err := config.Parse(withService(""))
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:3001", cfg.Services["api"].URL)
}

func TestParseKeepsAnEscapedPlaceholder(t *testing.T) {
	cfg, err := config.Parse(withService(`Plugins: {JWT: {Issuer: "$${NOT_A_VARIABLE} and $$5"}}`))
	require.NoError(t, err)

	assert.Equal(t, "${NOT_A_VARIABLE} and $$5", cfg.Plugins.JWTConfig.Issuer, "only $${ is an escape")
}

func TestParseLeavesKeysAndCommentsAlone(t *testing.T) {
	cfg, err := config.Parse(withService(`# Set ${UNSET_IN_A_COMMENT} before starting
`))
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	_, err = config.Parse(withService("${UNSET_KEY}: 1\n"))
	assert.ErrorContains(t, err, "field ${UNSET_KEY} not found", "a key is not filled, so it is an unknown key")
}

func TestParseReportsEveryPlaceholderThatCannotBeFilled(t *testing.T) {
	_, err := config.Parse([]byte(`
Plugins:
  JWT:
    SecretKey: ${UNSET_SECRET}
    Issuer: ${1BAD}
    Audience: ${UNCLOSED
Services:
  api:
    URL: ${UNSET_URL}
    Endpoints: [{HTTPMethod: GET, Path: /items}]
`))

	require.Error(t, err)
	for _, want := range []string{
		"the config has placeholders that cannot be filled",
		"line 4: ${UNSET_SECRET}: UNSET_SECRET is not set; set it, or give a fallback with ${UNSET_SECRET:-fallback}",
		`line 5: ${1BAD}: "1BAD" is not a variable name`,
		"line 6: ${UNCLOSED has no closing }",
		"line 9: ${UNSET_URL}: UNSET_URL is not set",
	} {
		assert.ErrorContains(t, err, want)
	}
}

func TestParseReportsTheLineOfAValueOfTheWrongType(t *testing.T) {
	t.Setenv("GATEWAY_PORT", "four")

	_, err := config.Parse(withService("Server:\n  Port: ${GATEWAY_PORT}\n"))

	assert.ErrorContains(t, err, "line 2: cannot unmarshal !!str `four` into int")
}

func TestParseStillRejectsUnknownKeysNextToPlaceholders(t *testing.T) {
	_, err := config.Parse(withService("Server:\n  Port: ${GATEWAY_PORT:-4000}\n  Prot: 1\n"))

	assert.ErrorContains(t, err, "line 3: field Prot not found")
}
