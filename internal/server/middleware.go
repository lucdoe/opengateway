package server

import (
	"net/http"

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

type MiddlewareConfig struct {
	LoggerConfig    logger.LoggerConfig
	JWTConfig       auth.JWTConfig
	CacheConfig     cache.CacheConfig
	RateLimitConfig ratelimit.RateLimitConfig
	CORSConfig      cors.CORSConfig
}

func InitMiddleware(
	cfg MiddlewareConfig,
	log logger.Logger,
	authService auth.AuthInterface,
	cacheService cache.Cache,
	rateLimiter ratelimit.RateLimiter,
	corsMiddleware cors.CORS,
) (map[string]Middleware, error) {

	middlewares := map[string]Middleware{
		"logger":     mw.NewLoggingMiddleware(log),
		"auth":       mw.NewAuthMiddleware(authService),
		"cache":      mw.NewCacheMiddleware(cacheService, &mw.StandardResponseUtil{}),
		"rate-limit": mw.NewRateLimitMiddleware(rateLimiter),
		"cors":       mw.NewCORSMiddleware(corsMiddleware),
	}

	return middlewares, nil
}
