// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	allowedOrigin = "http://allowed.com"
	OPTIONS       = "OPTIONS"
)

type mockCors struct{}

func (m *mockCors) ValidateOrigin(origin string) bool {
	return origin == allowedOrigin
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
		{"Valid OPTIONS but Invalid Method", OPTIONS, allowedOrigin, "POST", "", http.StatusForbidden},
		{"Valid OPTIONS but Invalid Headers", OPTIONS, allowedOrigin, "GET", "Wrong-Header", http.StatusForbidden},
		{"Valid GET Request", "GET", allowedOrigin, "", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "http://example.com", nil)
			request.Header.Set("Origin", tt.origin)
			if tt.method == OPTIONS {
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
