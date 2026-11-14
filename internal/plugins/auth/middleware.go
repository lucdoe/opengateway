// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
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
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}
			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || token == "" || strings.Contains(token, " ") {
				http.Error(w, "Authorization header must be Bearer token", http.StatusUnauthorized)
				return
			}

			claims, err := v.Validate(token)
			switch {
			case errors.Is(err, ErrMissingScope):
				http.Error(w, "Token lacks a required scope", http.StatusForbidden)
				return
			case err != nil:
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
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
				http.Error(w, "Token lacks a required scope", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
