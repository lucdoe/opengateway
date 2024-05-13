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
	srv "github.com/lucdoe/open-gateway/gateway/internal/server"
)

func main() {
	cfg, err := config.NewParser("./cmd/gateway/config.yaml").Parse()
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	router := mux.NewRouter()
	proxyService := proxy.NewProxyService()

	middlewareConfig := srv.MiddlewareConfig{
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

	middlewares, err := srv.InitMiddleware(middlewareConfig, os.Stderr)
	if err != nil {
		log.Fatalf("Failed to initialize middlewares: %v", err)
	}

	server := srv.NewServer(cfg, router, proxyService, middlewares)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}
