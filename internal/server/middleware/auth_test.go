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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

const (
	validTokenStr = "valid-token"
)

type mockAuth struct{}

func (m *mockAuth) Validate(token string) (jwt.Claims, error) {
	if token == validTokenStr {
		claims := jwt.MapClaims{"user": "123", "role": "admin"}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (m *mockAuth) ParseToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	if tokenStr == validTokenStr {
		return &jwt.RegisteredClaims{
			Subject:  "123",
			Issuer:   "test_issuer",
			Audience: []string{"test_audience"},
		}, nil
	}
	return nil, errors.New("invalid token")
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		token              string
		expectedStatusCode int
	}{
		{"No Token", "", http.StatusUnauthorized},
		{"Invalid Token", "invalid-token", http.StatusUnauthorized},
		{"Valid Token", validTokenStr, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			mockAuthService := &mockAuth{}
			middleware := NewAuthMiddleware(mockAuthService).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			middleware.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatusCode {
				t.Errorf("Expected status %d, got %d", tt.expectedStatusCode, rec.Code)
			}
		})
	}
}
