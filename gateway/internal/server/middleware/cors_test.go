package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockCors struct{}

func (m *mockCors) ValidateOrigin(origin string) bool {
	return origin == "http://allowed.com"
}

func (m *mockCors) ValidateMethod(method string) bool {
	return method == "GET"
}

func (m *mockCors) ValidateHeaders(headers string) bool {
	return headers == "X-Custom-Header"
}

func (m *mockCors) GetAllowedMethods() string {
	return "GET, POST, PUT, DELETE"
}

func (m *mockCors) GetAllowedHeaders() string {
	return "X-Custom-Header, Content-Type"
}

func TestCORSMiddleware(t *testing.T) {
	corsObj := &mockCors{}
	middleware := NewCORSMiddleware(corsObj)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name               string
		method             string
		origin             string
		requestMethod      string
		requestHeaders     string
		expectedStatusCode int
	}{
		{"Invalid Origin", "GET", "http://forbidden.com", "", "", http.StatusForbidden},
		{"Valid OPTIONS but Invalid Method", "OPTIONS", "http://allowed.com", "POST", "", http.StatusForbidden},
		{"Valid OPTIONS but Invalid Headers", "OPTIONS", "http://allowed.com", "GET", "Wrong-Header", http.StatusForbidden},
		{"Valid GET Request", "GET", "http://allowed.com", "", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "http://example.com", nil)
			request.Header.Set("Origin", tt.origin)
			if tt.method == "OPTIONS" {
				request.Header.Set("Access-Control-Request-Method", tt.requestMethod)
				request.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)
			}

			recorder := httptest.NewRecorder()
			handler := middleware.Middleware(testHandler)
			handler.ServeHTTP(recorder, request)

			if recorder.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, recorder.Code)
			}
		})
	}
}
