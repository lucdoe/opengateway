package server

import (
	"fmt"
	"io"
	"net/http"
	"time"

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
	LoggerConfig      logger.LoggerConfig
	JWTConfig         auth.JWTConfig
	RedisAddr         string
	RedisPassword     string
	RateLimitCapacity int64
	RateLimitWindow   time.Duration
	CORSConfig        cors.CORSConfig
}

func InitMiddleware(cfg MiddlewareConfig, errorOutput io.Writer) (map[string]Middleware, error) {
	logCfg := cfg.LoggerConfig
	logCfg.ErrOutput = errorOutput
	log, err := logger.NewLogger(logCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	authService, err := auth.NewAuthService(cfg.JWTConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AuthService: %w", err)
	}

	cacheService := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword)
	rateLimiter := ratelimit.NewRateLimitService(cacheService, cfg.RateLimitCapacity, cfg.RateLimitWindow)
	corsMiddleware := cors.NewCors(cfg.CORSConfig)

	middlewares := map[string]Middleware{
		"logger":     mw.NewLoggingMiddleware(log),
		"auth":       mw.NewAuthMiddleware(authService),
		"cache":      mw.NewCacheMiddleware(cacheService, &mw.StandardResponseUtil{}),
		"rate-limit": mw.NewRateLimitMiddleware(rateLimiter),
		"cors":       mw.NewCORSMiddleware(corsMiddleware),
	}

	return middlewares, nil
}
