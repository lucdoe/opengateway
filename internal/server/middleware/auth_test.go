package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	mw "github.com/lucdoe/open-gateway/gateway/internal/server/middleware"
)

type mockAuth struct{}

func (m *mockAuth) Validate(token string) (jwt.Claims, error) {
	if token == "valid-token" {
		claims := jwt.MapClaims{"user": "123", "role": "admin"}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (m *mockAuth) ParseToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	if tokenStr == "valid-token" {
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
		{"Valid Token", "valid-token", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			mockAuthService := &mockAuth{}
			middleware := mw.NewAuthMiddleware(mockAuthService).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			middleware.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatusCode {
				t.Errorf("Expected status %d, got %d", tt.expectedStatusCode, rec.Code)
			}
		})
	}
}
