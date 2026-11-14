// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/lucdoe/opengateway/internal/httpx"
)

// Validator checks a bearer token and returns its claims.
type Validator interface {
	Validate(token string) (*Claims, error)
}

type claimsKey struct{}

func claimsFrom(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey{}).(*Claims)
	return claims
}

// NewMiddleware returns the auth plugin. It answers 401 to requests without
// a Bearer token the validator accepts, and 403 to tokens that lack a
// required scope. Accepted claims go on the request context for RequireScope.
func NewMiddleware(v Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpx.Error(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}
			scheme, token, _ := strings.Cut(authHeader, " ")
			if !strings.EqualFold(scheme, "Bearer") || token == "" || strings.Contains(token, " ") {
				httpx.Error(w, http.StatusUnauthorized, "Authorization header must be Bearer token")
				return
			}

			claims, err := v.Validate(token)
			switch {
			case errors.Is(err, ErrMissingScope):
				httpx.Error(w, http.StatusForbidden, "Token lacks a required scope")
				return
			case err != nil:
				// Why it failed would help someone forging tokens, not the caller.
				httpx.Error(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims)))
		})
	}
}

// RequireScope answers 403 unless the token the auth plugin accepted carries
// every space-separated scope in scope. It must run after the auth plugin.
func RequireScope(scope string) func(http.Handler) http.Handler {
	required := strings.Fields(scope)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if claims := claimsFrom(r.Context()); claims == nil || !claims.HasScopes(required) {
				httpx.Error(w, http.StatusForbidden, "Token lacks a required scope")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
