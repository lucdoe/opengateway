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

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "github.com/lucdoe/opengateway/internal/server/middleware"
)

type mockLogger struct {
	Messages []string
	Done     chan bool
}

func (m *mockLogger) Info(msg string, details string) {
	m.Messages = append(m.Messages, msg+" "+details)
	m.Done <- true
}

func TestLoggingMiddleware(t *testing.T) {
	mockLog := &mockLogger{
		Done: make(chan bool, 1),
	}
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

			select {
			case <-mockLog.Done:
				if len(mockLog.Messages) == 0 || mockLog.Messages[0] != tt.expectedLog {
					t.Errorf("Expected log '%s', got '%v'", tt.expectedLog, mockLog.Messages)
				}
			case <-time.After(1 * time.Second):
				t.Error("Expected log message but got none")
			}
		})
	}
}
