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
	"net/http"

	"github.com/lucdoe/opengateway/internal/plugins/auth"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
	"github.com/lucdoe/opengateway/internal/plugins/cors"
	"github.com/lucdoe/opengateway/internal/plugins/logger"
	ratelimit "github.com/lucdoe/opengateway/internal/plugins/rate-limit"
	mw "github.com/lucdoe/opengateway/internal/server/middleware"
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
