// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
