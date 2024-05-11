package server

import (
	"net/http"
	"time"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
	ratelimit "github.com/lucdoe/open-gateway/gateway/internal/plugins/rate-limit"
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

	middlewares := map[string]Middleware{
		"logger":     logger,
		"cache":      cache,
		"rate-limit": rateLimiter,
	}
	return middlewares, nil
}
