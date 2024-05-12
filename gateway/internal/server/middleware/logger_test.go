package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/lucdoe/open-gateway/gateway/internal/server/middleware"
)

type mockLogger struct {
	Messages []string
}

func (m *mockLogger) Info(msg string, details string) {
	m.Messages = append(m.Messages, msg+" "+details)
}

func TestLoggingMiddleware(t *testing.T) {
	mockLog := &mockLogger{}
	middleware := mw.NewLoggingMiddleware(mockLog)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name        string
		method      string
		path        string
		remoteAddr  string
		expectedLog string
	}{
		{
			name:        "Logging GET Request",
			method:      "GET",
			path:        "/test",
			remoteAddr:  "127.0.0.1",
			expectedLog: "Request GET /test from 127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "http://example.com"+tt.path, nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			handler := middleware.Middleware(testHandler)
			handler.ServeHTTP(rec, req)

			if len(mockLog.Messages) == 0 || mockLog.Messages[0] != tt.expectedLog {
				t.Errorf("Expected log '%s', got '%v'", tt.expectedLog, mockLog.Messages)
			}
		})
	}
}
