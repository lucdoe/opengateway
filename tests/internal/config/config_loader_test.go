package config_test

import (
	"errors"
	"testing"

	"github.com/lucdoe/opengateway/internal/config"
)

type MockFileReader struct {
	Data []byte
	Err  error
}

func (mfr *MockFileReader) ReadFile(filePath string) ([]byte, error) {
	return mfr.Data, mfr.Err
}

func TestLoaderLoadConfigFromFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockData := []byte("title: Example\nversion: 1.0")
		loader := config.NewLoader(&MockFileReader{Data: mockData, Err: nil})
		config, err := loader.LoadConfigFromFile("does_not_matter.yaml")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if config.Title != "Example" {
			t.Errorf("Expected title to be 'Example', got %s", config.Title)
		}

		if config.Version != "1.0" {
			t.Errorf("Expected version to be '1.0', got %s", config.Version)
		}
	})

	t.Run("read file error", func(t *testing.T) {
		loader := config.NewLoader(&MockFileReader{Data: nil, Err: errors.New("failed to read file")})
		_, err := loader.LoadConfigFromFile("does_not_matter.yaml")

		if err == nil {
			t.Error("Expected an error, got none")
		}
	})
}
