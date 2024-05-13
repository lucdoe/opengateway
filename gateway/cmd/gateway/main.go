package main

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cors"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
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

type MiddlewareInitializer interface {
	InitMiddleware(cfg server.MiddlewareConfig) (map[string]server.Middleware, error)
}

type DefaultMiddlewareInitializer struct{}

func (d *DefaultMiddlewareInitializer) InitMiddleware(cfg server.MiddlewareConfig) (map[string]server.Middleware, error) {
	return server.InitMiddleware(cfg, os.Stderr)
}

type ServerDependencies struct {
	ConfigLoader          ConfigLoader
	MiddlewareInitializer MiddlewareInitializer
	Router                *mux.Router
	ProxyService          proxy.ProxyService
}

func InitializeServer(deps ServerDependencies) (*server.Server, error) {
	cfg, err := deps.ConfigLoader.LoadConfig("./cmd/gateway/config.yaml")
	if err != nil {
		return nil, err
	}

	middlewares, err := deps.MiddlewareInitializer.InitMiddleware(middlewareConfig(cfg))
	if err != nil {
		return nil, err
	}

	return server.NewServer(cfg, deps.Router, deps.ProxyService, middlewares), nil
}

func middlewareConfig(cfg *config.Config) server.MiddlewareConfig {
	return server.MiddlewareConfig{
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
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		RateLimitCapacity: 100,
		RateLimitWindow:   1 * time.Minute,
		CORSConfig: cors.CORSConfig{
			Origins: "*",
			Methods: "GET, POST, PUT, DELETE, OPTIONS",
			Headers: "Content-Type, Authorization",
		},
	}
}

func main() {
	deps := ServerDependencies{
		ConfigLoader:          &DefaultConfigLoader{},
		MiddlewareInitializer: &DefaultMiddlewareInitializer{},
		Router:                mux.NewRouter(),
		ProxyService:          proxy.NewProxyService(),
	}

	server, err := InitializeServer(deps)
	if err != nil {
		log.Fatalf("Failed to initialize the server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}
