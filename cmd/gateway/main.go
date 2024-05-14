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

package main

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cors"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
	ratelimit "github.com/lucdoe/open-gateway/gateway/internal/plugins/rate-limit"
	"github.com/lucdoe/open-gateway/gateway/internal/proxy"
	"github.com/lucdoe/open-gateway/gateway/internal/server"
)

type ConfigLoader interface {
	LoadConfig(path string) (*config.Config, error)
}

type DefaultConfigLoader struct{}

func (d *DefaultConfigLoader) LoadConfig(path string) (*config.Config, error) {
	return config.NewParser(path).Parse()
}

type ServerDependencies struct {
	ConfigLoader          ConfigLoader
	MiddlewareInitializer server.MiddlewareConfig
	Router                *mux.Router
	ProxyService          proxy.ProxyService
	CacheService          cache.Cache
}

func InitializeServer(deps ServerDependencies) (*server.Server, error) {
	cfg, err := deps.ConfigLoader.LoadConfig("./cmd/gateway/config.yaml")
	if err != nil {
		return nil, err
	}

	cacheService := deps.CacheService

	rateLimiter := ratelimit.NewRateLimitService(ratelimit.RateLimitConfig{
		Store:  cacheService,
		Limit:  100,
		Window: 1 * time.Minute,
	})

	middlewareCfg := server.MiddlewareConfig{
		LoggerConfig: logger.LoggerConfig{
			FilePath:     "server.log",
			FileWriter:   nil,
			ErrOutput:    os.Stderr,
			TimeProvider: logger.RealTime{},
			FileOpener:   logger.DefaultFileOpener{},
		},
		JWTConfig: auth.JWTConfig{
			SecretKey:     []byte("your-256-bit-secret"),
			SigningMethod: jwt.SigningMethodHS256,
			Issuer:        "ExampleIssuer",
			Audience:      "ExampleAudience",
			Scope:         "ExampleScope",
		},
		CacheConfig: cache.CacheConfig{
			Addr:     "localhost:6379",
			Password: "",
		},
		RateLimitConfig: ratelimit.RateLimitConfig{
			Store:  cacheService,
			Limit:  100,
			Window: 1 * time.Minute,
		},
		CORSConfig: cors.CORSConfig{
			Origins: "*",
			Methods: "GET, POST, PUT, DELETE, OPTIONS",
			Headers: "Content-Type, Authorization",
		},
	}

	middlewares, err := server.InitMiddleware(middlewareCfg, logger.NewLogger(middlewareCfg.LoggerConfig), auth.NewAuthService(middlewareCfg.JWTConfig), cacheService, rateLimiter, cors.NewCors(middlewareCfg.CORSConfig))
	if err != nil {
		return nil, err
	}

	return server.NewServer(cfg, deps.Router, deps.ProxyService, middlewares), nil
}

func main() {
	deps := ServerDependencies{
		ConfigLoader: &DefaultConfigLoader{},
		Router:       mux.NewRouter(),
		ProxyService: proxy.NewProxyService(),
		CacheService: cache.NewRedisCache(cache.CacheConfig{
			Addr:     "localhost:6379",
			Password: "",
		}),
	}

	server, err := InitializeServer(deps)
	if err != nil {
		log.Fatalf("Failed to initialize the server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}