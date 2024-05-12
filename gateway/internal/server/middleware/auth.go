package middleware

import (
	"net/http"
	"strings"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
)

type AuthMiddleware struct {
	AuthService *auth.Auth
}

func NewAuthMiddleware(as *auth.Auth) *AuthMiddleware {
	return &AuthMiddleware{
		AuthService: as,
	}
}

func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header must be Bearer token", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		_, err := am.AuthService.Validate(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
