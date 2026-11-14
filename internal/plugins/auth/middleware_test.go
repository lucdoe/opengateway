// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// errAny marks a test case that expects some error other than ErrMissingScope.
var errAny = errors.New("any error")

type fakeValidator struct{}

func (fakeValidator) Validate(token string) (*Claims, error) {
	switch token {
	case "valid":
		return &Claims{Scope: "read"}, nil
	case "unscoped":
		return nil, ErrMissingScope
	}
	return nil, errors.New("bad signature")
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		header string
		status int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"not a bearer token", "Basic abc", http.StatusUnauthorized},
		{"invalid token", "Bearer nope", http.StatusUnauthorized},
		{"token without the required scope", "Bearer unscoped", http.StatusForbidden},
		{"valid token", "Bearer valid", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var seen *Claims
			handler := NewMiddleware(fakeValidator{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = claimsFrom(r.Context())
			}))
			req := httptest.NewRequest("GET", "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
			if tc.status == http.StatusOK {
				assert.Equal(t, "read", seen.Scope, "claims reach the next handler")
			}
		})
	}
}

func TestRequireScope(t *testing.T) {
	tests := []struct {
		name   string
		claims *Claims
		status int
	}{
		{"has every scope", &Claims{Scope: "read write"}, http.StatusOK},
		{"lacks one", &Claims{Scope: "read"}, http.StatusForbidden},
		{"auth did not run", nil, http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := RequireScope("write read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			req := httptest.NewRequest("GET", "/", nil)
			if tc.claims != nil {
				req = req.WithContext(context.WithValue(req.Context(), claimsKey{}, tc.claims))
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
		})
	}
}
