package logger_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
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

	request, _ := http.NewRequest("GET", "/test", nil)
	request.RemoteAddr = "123.123.123.123"

	logger.Info("A simple request", request)

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

func TestApplyMiddleware(t *testing.T) {
	mockFile := new(MockFileWriter)
	log, err := logger.NewLogger("", mockFile)
	if err != nil {
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

	middleware := log.Middleware(mockHandler)

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
