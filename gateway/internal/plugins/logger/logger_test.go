package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
)

type MockFileWriter struct {
	bytes.Buffer
}

func (m *MockFileWriter) Write(p []byte) (n int, err error) {
	return m.Buffer.Write(p)
}

func TestOSLogger(t *testing.T) {
	mockFile := new(MockFileWriter)
	logger, err := logger.NewLogger("", mockFile)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("A simple request", "GET /test from 123.123.123.123")

	expectedParts := []string{
		"INFO",
		"A simple request",
		"GET /test",
		"from 123.123.123.123",
	}

	logOutput := mockFile.String()
	for _, part := range expectedParts {
		if !strings.Contains(logOutput, part) {
			t.Errorf("Log output does not contain expected part: %s. Full log: %s", part, logOutput)
		}
	}
}
