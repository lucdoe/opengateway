package server

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cors"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
	ratelimit "github.com/lucdoe/open-gateway/gateway/internal/plugins/rate-limit"
	mw "github.com/lucdoe/open-gateway/gateway/internal/server/middleware"
)

type Middleware interface {
	Middleware(next http.Handler) http.Handler
}

func InitMiddleware() (map[string]Middleware, error) {
	logFileOpener := logger.DefaultFileOpener{}
	timeProvider := logger.RealTime{}
	errorOutput := os.Stderr

	logger, err := logger.NewLogger("server.log", nil, errorOutput, timeProvider, logFileOpener)
	if err != nil {
		return nil, err
	}

	config := auth.JWTConfig{
		SecretKey:     []byte("your-256-bit-secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        "ExampleIssuer",
		Audience:      "ExampleAudience",
		Scope:         "ExampleScope",
	}
	auth, err := auth.NewAuthService(config)
	if err != nil {
		log.Fatalf("Failed to create AuthService: %v", err)
	}

	cache := cache.NewRedisCache("localhost:6379", "")
	rateLimiter := ratelimit.NewRateLimitService(cache, 60, time.Minute)

	corsConfig := cors.CORSConfig{
		Origins: "*",
		Methods: "GET, POST, PUT, DELETE, OPTIONS",
		Headers: "Content-Type, Authorization",
	}
	corsMiddleware := cors.NewCors(corsConfig)

	middlewares := map[string]Middleware{
		"logger":     mw.NewLoggingMiddleware(logger),
		"cache":      mw.NewCacheMiddleware(cache),
		"rate-limit": mw.NewRateLimitMiddleware(rateLimiter),
		"cors":       mw.NewCORSMiddleware(corsMiddleware),
		"auth":       mw.NewAuthMiddleware(auth),
	}

	return middlewares, nil
}
