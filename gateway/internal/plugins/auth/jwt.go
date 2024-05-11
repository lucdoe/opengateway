package pluginJWT

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware interface {
	Middleware() func(http.Handler) http.Handler
	Validate(tokenStr string) (jwt.Claims, error)
}

type JWTConfig struct {
	SecretKey     []byte
	SigningMethod jwt.SigningMethod
	Issuer        string
	Audience      string
	Scope         string
}

type JWTService struct {
	config JWTConfig
}

func NewJWTService(config JWTConfig) AuthMiddleware {
	return &JWTService{
		config: config,
	}
}

func (j *JWTService) Validate(tokenStr string) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.config.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok && !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (j *JWTService) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
			_, err := j.Validate(tokenStr)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
