// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestResolveConfigPath(t *testing.T) {
	t.Run("the -c flag wins", func(t *testing.T) {
		t.Setenv("CONFIG_PATH", "env.yaml")
		assert.Equal(t, "flag.yaml", resolveConfigPath("flag.yaml"))
	})
	t.Run("then CONFIG_PATH", func(t *testing.T) {
		t.Setenv("CONFIG_PATH", "env.yaml")
		assert.Equal(t, "env.yaml", resolveConfigPath(""))
	})
	t.Run("then the default", func(t *testing.T) {
		t.Setenv("CONFIG_PATH", "")
		assert.Equal(t, DefaultConfigPath, resolveConfigPath(""))
	})
}

// runCLI runs the command with the real config loader and returns its exit
// code and output.
func runCLI(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run(args, ServerDependencies{loadConfig: config.Load}, &out, &errOut)
	return code, out.String(), errOut.String()
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCheckAcceptsTheShippedConfig(t *testing.T) {
	code, stdout, stderr := runCLI("check", "-c", "config.yaml")

	assert.Equal(t, 0, code, stderr)
	assert.Equal(t, "config.yaml is valid: 2 services, 7 endpoints\n", stdout)
}

func TestCheckReadsCONFIG_PATH(t *testing.T) {
	t.Setenv("CONFIG_PATH", "config.yaml")
	code, stdout, _ := runCLI("check")

	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "config.yaml is valid")
}

func TestCheckReportsEveryProblem(t *testing.T) {
	path := writeConfig(t, `
Services:
  users:
    URL: "ftp://users"
    Endpoints:
      - {HTTPMethod: FETCH, Path: /users}
`)
	code, stdout, stderr := runCLI("check", "-c", path)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, `URL "ftp://users" must be an absolute http or https URL`)
	assert.Contains(t, stderr, `HTTPMethod "FETCH" is not one of`)
}

func TestCheckPrintsWarningsButPasses(t *testing.T) {
	path := writeConfig(t, `
Plugins: {JWTConfig: {SecretKey: your-256-bit-secret, Issuer: auth, Audience: gateway}}
Services:
  users:
    URL: http://users.example.com
    Endpoints: [{HTTPMethod: GET, Path: /me, Auth: {ApplyAuth: true}}]
`)
	code, stdout, stderr := runCLI("check", "-c", path)

	assert.Equal(t, 0, code, "warnings do not fail the check")
	assert.Contains(t, stdout, "is valid: 1 services, 1 endpoints")
	assert.Equal(t, "warning: Plugins.JWTConfig.SecretKey is the example secret from the docs, so anyone can sign tokens with it; set your own, for example with JWT_SECRET\n"+
		"warning: service \"users\": tokens are forwarded over plain http to users.example.com; use https\n", stderr)
}

func TestCheckReportsAMissingFile(t *testing.T) {
	code, _, stderr := runCLI("check", "-c", filepath.Join(t.TempDir(), "missing.yaml"))

	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, "reading config")
}

func TestRunRejectsBadUsage(t *testing.T) {
	for name, args := range map[string][]string{
		"an unknown flag":      {"check", "-x"},
		"a flag without value": {"check", "-c"},
		"an unknown command":   {"chek"},
		"an extra argument":    {"check", "-c", "config.yaml", "more"},
	} {
		t.Run(name, func(t *testing.T) {
			code, stdout, stderr := runCLI(args...)
			assert.Equal(t, 2, code)
			assert.Empty(t, stdout)
			assert.Contains(t, stderr, "Usage:")
		})
	}
}

func TestRunPrintsHelp(t *testing.T) {
	code, _, stderr := runCLI("-h")

	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "opengateway check [-c config.yaml]")
}

func TestServeFailsOnAnInvalidConfig(t *testing.T) {
	path := writeConfig(t, "Services: {}\n")
	code, _, _ := runCLI("-c", path)

	assert.Equal(t, 1, code)
}
