package server

import (
	"net/http"
	"time"

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
	logger, err := logger.NewLogger("server.log", nil)
	if err != nil {
		return nil, err
	}

	cache := cache.NewCache("localhost:6379", "")
	rateLimiter := ratelimit.NewRateLimitService(cache, 60, time.Minute)

	corsConfig := cors.CORSConfig{
		Origins: "*",
		Methods: "GET, POST, PUT, DELETE, OPTIONS",
		Headers: "Content-Type, Authorization",
	}
	corsMiddleware := cors.NewCors(corsConfig)

	middlewares := map[string]Middleware{
		"logger":     logger,
		"cache":      cache,
		"rate-limit": mw.NewRateLimitMiddleware(rateLimiter),
		"cors":       corsMiddleware,
	}

	return middlewares, nil
}
