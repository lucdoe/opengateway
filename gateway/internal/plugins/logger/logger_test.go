package logger_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	log := logger.NewLogger("test.log", mockFile)

	if err := log.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	request, _ := http.NewRequest("GET", "/test", nil)
	request.RemoteAddr = "123.123.123.123"
	duration := 50 * time.Millisecond

	log.Info("A simple request", request, duration)

	expectedParts := []string{
		"INFO",
		"A simple request",
		"GET /test",
		"from 123.123.123.123",
		"in 50ms",
	}

	logOutput := mockFile.String()
	for _, part := range expectedParts {
		if !strings.Contains(logOutput, part) {
			t.Errorf("Log output does not contain expected part: %s", part)
		}
	}
}

func TestOSLoggerError(t *testing.T) {
	mockFile := new(MockFileWriter)
	log := logger.NewLogger("test.log", mockFile)

	if err := log.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	log.Error("A simple error")

	expectedParts := "[ERROR]: A simple error"

	logOutput := mockFile.String()
	fmt.Println(logOutput)
	if !strings.Contains(logOutput, expectedParts) {
		t.Errorf("Log output does not contain expected part: %s", expectedParts)
	}
}

func TestApplyMiddleware(t *testing.T) {
	mockFile := new(MockFileWriter)
	log := logger.NewLogger("", mockFile)
	if err := log.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Could not create HTTP request: %v", err)
	}

	recorder := httptest.NewRecorder()

	middleware := log.Apply(mockHandler)

	middleware.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	logOutput := mockFile.String()
	expectedParts := []string{
		"INFO",
		"Request",
		"/test",
		"GET",
	}
	for _, part := range expectedParts {
		if !strings.Contains(logOutput, part) {
			t.Errorf("Log output does not contain expected part '%s'. Full log: %s", part, logOutput)
		}
	}
}
