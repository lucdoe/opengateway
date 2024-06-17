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
	"testing"

	parser "github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	yamlData := `
Services:
  service1:
    URL: "http://example.com"
    Protocol: "HTTP"
    Endpoints:
      - Name: "Endpoint1"
        HTTPMethod: "GET"
        Path: "/data"
`
	var config parser.Config
	p := parser.NewParser("")
	err := p.Unmarshal([]byte(yamlData), &config)
	assert.NoError(t, err)

	require.Len(t, config.Services, 1)
	assert.Equal(t, "http://example.com", config.Services["service1"].URL)
	assert.Equal(t, "HTTP", config.Services["service1"].Protocol)
	assert.Len(t, config.Services["service1"].Endpoints, 1)
	assert.Equal(t, "Endpoint1", config.Services["service1"].Endpoints[0].Name)
	assert.Equal(t, "GET", config.Services["service1"].Endpoints[0].HTTPMethod)
	assert.Equal(t, "/data", config.Services["service1"].Endpoints[0].Path)
}

func TestUnmarshalWithOptionalFields(t *testing.T) {
	yamlData := `
Services:
  service2:
    URL: "http://example.org"
    Protocol: "HTTPS"
    Subpath: "/api"
    Endpoints:
      - Name: "Endpoint2"
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
	var config parser.Config
	p := parser.NewParser("")
	err := p.Unmarshal([]byte(yamlData), &config)
	assert.NoError(t, err)

	require.Len(t, config.Services, 1)
	assert.Equal(t, "http://example.org", config.Services["service2"].URL)
	assert.Equal(t, "HTTPS", config.Services["service2"].Protocol)
	assert.Equal(t, "/api", config.Services["service2"].Subpath)
	assert.Len(t, config.Services["service2"].Endpoints, 1)
	assert.Equal(t, "Endpoint2", config.Services["service2"].Endpoints[0].Name)
	assert.Equal(t, "POST", config.Services["service2"].Endpoints[0].HTTPMethod)
	assert.Equal(t, "/submit", config.Services["service2"].Endpoints[0].Path)
	assert.Len(t, config.Services["service2"].Endpoints[0].QueryParams, 1)
	assert.Equal(t, "key", config.Services["service2"].Endpoints[0].QueryParams[0].Key)
	assert.Equal(t, "value", config.Services["service2"].Endpoints[0].QueryParams[0].Value)
	assert.True(t, config.Services["service2"].Endpoints[0].Auth.ApplyAuth)
	assert.Equal(t, "Bearer", config.Services["service2"].Endpoints[0].Auth.Method)
	assert.Equal(t, "HS256", config.Services["service2"].Endpoints[0].Auth.Algorithm)
	assert.Equal(t, "read:write", config.Services["service2"].Endpoints[0].Auth.Scope)
	assert.Equal(t, "secret", config.Services["service2"].Endpoints[0].Auth.SecretKey)
	assert.Equal(t, map[string]interface{}{"key1": "value1"}, config.Services["service2"].Endpoints[0].Body)
	assert.Len(t, config.Services["service2"].Endpoints[0].Plugins, 1)
	assert.Equal(t, "logger", config.Services["service2"].Endpoints[0].Plugins[0])
}

func TestParse(t *testing.T) {
	testFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile.Name())

	yamlContent := `
Services:
  service1:
    URL: "http://example.com"
    Protocol: "HTTP"
    Endpoints:
      - Name: "Endpoint1"
        HTTPMethod: "GET"
        Path: "/data"
`
	_, err = testFile.WriteString(yamlContent)
	require.NoError(t, err)
	testFile.Close()

	p := parser.NewParser(testFile.Name())

	config, err := p.Parse()
	assert.NoError(t, err)
	require.Len(t, config.Services, 1)

	service, exists := config.Services["service1"]
	require.True(t, exists)
	assert.Equal(t, "http://example.com", service.URL)
	assert.Equal(t, "HTTP", service.Protocol)
	assert.Len(t, service.Endpoints, 1)
	assert.Equal(t, "Endpoint1", service.Endpoints[0].Name)
	assert.Equal(t, "GET", service.Endpoints[0].HTTPMethod)
	assert.Equal(t, "/data", service.Endpoints[0].Path)
}

func TestParseWithInvalidYAML(t *testing.T) {
	testFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile.Name())

	invalidYamlContent := `
Services:
  - URL: "http://example.com"
    Protocol: "HTTP"
    Endpoints:
      - Name: "Endpoint1"
        HTTPMethod: "GET"
        Path: "/data"
`
	_, err = testFile.WriteString(invalidYamlContent)
	require.NoError(t, err)
	testFile.Close()

	p := parser.NewParser(testFile.Name())

	_, err = p.Parse()
	assert.Error(t, err)
}

func TestReadFileError(t *testing.T) {
	p := parser.NewParser("nonexistent.yaml")

	_, err := p.Parse()
	assert.Error(t, err)
}
