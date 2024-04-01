package main__test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIfDirectoriesExist(t *testing.T) {
	rootDir := "./"

	directories := []string{
		"cmd/fire_gateway",
		"deployment",
		"docs",
		"internal/config",
		"internal/middleware",
		"internal/data",
		"internal/http",
		"internal/logger",
		"logs",
	}

	for _, dir := range directories {
		fullPath := filepath.Join(rootDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Directory /%s does not exist", dir)
		}
	}
}
