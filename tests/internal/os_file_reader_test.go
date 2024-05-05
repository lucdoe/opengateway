package utils

import (
	"os"
	"testing"

	"github.com/lucdoe/capstone_gateway/internal"
)

func TestReadFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some content to the file
	content := []byte("Hello, world!")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatalf("Cannot write to temporary file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Cannot close temporary file: %v", err)
	}

	reader := internal.OSFileReader{}
	readContent, err := reader.ReadFile(tmpfile.Name())
	if err != nil {
		t.Errorf("ReadFile returned an error: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("ReadFile = %q, want %q", readContent, content)
	}

	// Test error handling with non-existent file
	_, err = reader.ReadFile("nonexistentfile.txt")
	if err == nil {
		t.Errorf("Expected an error for nonexistent file, but got none")
	}
}
