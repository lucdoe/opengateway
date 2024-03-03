package tests

import (
	"os"
	"testing"
)

func TestDockerfileExists(t *testing.T) {
	if _, err := os.Stat("../deployments/Dockerfile"); os.IsNotExist(err) {
		t.Error("Dockerfile does not exist")
	}
}

func TestDockerIgnoreExists(t *testing.T) {
	if _, err := os.Stat("../deployments/.dockerignore"); os.IsNotExist(err) {
		t.Error(".dockerignore does not exist")
	}
}

func TestDockerComposeExists(t *testing.T) {
	if _, err := os.Stat("../deployments/docker-compose.yaml"); os.IsNotExist(err) {
		t.Error("docker-compose.yml does not exist")
	}
}
