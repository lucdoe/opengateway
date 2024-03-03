package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIfDirectoriesExist(t *testing.T) {
	rootDir := "../"

	directories := []string{
		"cmd/opengateway",
		"deployments",
		"docs",
		"internal/config",
		"internal/router",
		"internal/middleware",
		"internal/proxy",
		"internal/service",
		"internal/utils",
		"logs",
		"tests",
	}

	for _, dir := range directories {
		fullPath := filepath.Join(rootDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Directory /%s does not exist", dir)
		}
	}
}
