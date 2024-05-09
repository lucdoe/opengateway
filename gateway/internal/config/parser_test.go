package config_test

import (
	"os"
	"testing"

	parser "github.com/lucdoe/open-gateway/gateway/internal/config"
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
	parser := parser.NewParser("")
	err := parser.Unmarshal([]byte(yamlData), &config)
	assert.NoError(t, err)

	require.Len(t, config.Services, 1)
	assert.Equal(t, "http://example.com", config.Services["service1"].URL)
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

	parser := parser.NewParser((testFile.Name()))

	config, err := parser.Parse()
	assert.NoError(t, err)
	require.Len(t, config.Services, 1)

	service, exists := config.Services["service1"]
	require.True(t, exists)
	assert.Equal(t, "http://example.com", service.URL)
}
