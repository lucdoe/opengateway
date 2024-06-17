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

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/plugins/auth"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
	"github.com/lucdoe/opengateway/internal/plugins/cors"
	"github.com/lucdoe/opengateway/internal/plugins/logger"
	ratelimit "github.com/lucdoe/opengateway/internal/plugins/rate-limit"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/lucdoe/opengateway/internal/server"
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
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./cmd/opengateway/config.yaml"
	}
	cfg, err := deps.ConfigLoader.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	cacheService := deps.CacheService

	loggerConfig := cfg.Plugins.LoggerConfig
	jwtConfig := cfg.Plugins.JWTConfig
	rateLimitConfig := cfg.Plugins.RateLimitConfig
	corsConfig := cfg.Plugins.CORSConfig

	rateLimiter := ratelimit.NewRateLimitService(ratelimit.RateLimitConfig{
		Store:  cacheService,
		Limit:  rateLimitConfig.Limit,
		Window: rateLimitConfig.Window,
	})

	middlewareCfg := server.MiddlewareConfig{
		LoggerConfig: logger.LoggerConfig{
			FilePath:     loggerConfig.FilePath,
			FileWriter:   nil,
			ErrOutput:    os.Stderr,
			TimeProvider: logger.RealTime{},
			FileOpener:   logger.DefaultFileOpener{},
		},
		JWTConfig: auth.JWTConfig{
			SecretKey:     []byte(jwtConfig.SecretKey),
			SigningMethod: jwt.SigningMethodHS256,
			Issuer:        jwtConfig.Issuer,
			Audience:      jwtConfig.Audience,
			Scope:         jwtConfig.Scope,
		},
		RateLimitConfig: ratelimit.RateLimitConfig{
			Store:  cacheService,
			Limit:  rateLimitConfig.Limit,
			Window: rateLimitConfig.Window,
		},
		CORSConfig: cors.CORSConfig{
			Origins: corsConfig.Origins,
			Methods: corsConfig.Methods,
			Headers: corsConfig.Headers,
		},
	}

	middlewares, err := server.InitMiddleware(middlewareCfg, logger.NewLogger(middlewareCfg.LoggerConfig), auth.NewAuthService(middlewareCfg.JWTConfig), cacheService, rateLimiter, cors.NewCors(middlewareCfg.CORSConfig))
	if err != nil {
		return nil, err
	}

	return server.NewServer(cfg, deps.Router, deps.ProxyService, middlewares), nil
}

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisHost == "" || redisPort == "" {
		redisHost = "localhost"
		redisPort = "6379"
	}
	deps := ServerDependencies{
		ConfigLoader: &DefaultConfigLoader{},
		Router:       mux.NewRouter(),
		ProxyService: proxy.NewProxyService(),
		CacheService: cache.NewRedisCache(cache.CacheConfig{
			Addr:     redisHost + ":" + redisPort,
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
