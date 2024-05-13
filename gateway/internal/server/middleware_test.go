package server

import (
	"testing"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cors"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
	ratelimit "github.com/lucdoe/open-gateway/gateway/internal/plugins/rate-limit"
	mw "github.com/lucdoe/open-gateway/gateway/internal/server/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
	logger.Logger
}

func (m *MockLogger) Log(message string) {
	m.Called(message)
}

type MockAuthService struct {
	mock.Mock
	auth.Auth
}

type MockCacheService struct {
	mock.Mock
	cache.Cache
}

type MockRateLimiter struct {
	mock.Mock
	ratelimit.RateLimitConfig
}

type MockCORSMiddleware struct {
	mock.Mock
	cors.CORS
}

func TestInitMiddleware(t *testing.T) {
	mockLogger := new(MockLogger)
	mockAuthService := new(MockAuthService)
	mockCacheService := new(MockCacheService)
	mockRateLimiter := new(MockRateLimiter)
	mockCORSMiddleware := new(MockCORSMiddleware)

	config := MiddlewareConfig{
		LoggerConfig:    logger.LoggerConfig{},
		JWTConfig:       auth.JWTConfig{},
		CacheConfig:     cache.CacheConfig{},
		RateLimitConfig: ratelimit.RateLimitConfig{},
		CORSConfig:      cors.CORSConfig{},
	}

	middlewares, err := InitMiddleware(config, mockLogger, mockAuthService, mockCacheService, mockRateLimiter, mockCORSMiddleware)

	assert.NoError(t, err)
	assert.NotNil(t, middlewares)
	assert.IsType(t, &mw.LoggingMiddleware{}, middlewares["logger"])
	assert.IsType(t, &mw.AuthMiddleware{}, middlewares["auth"])
	assert.IsType(t, &mw.CacheMiddleware{}, middlewares["cache"])
	assert.IsType(t, &mw.RateLimitMiddleware{}, middlewares["rate-limit"])
	assert.IsType(t, &mw.CORSMiddleware{}, middlewares["cors"])
}
