package main

import (
	"log"

	"github.com/gorilla/mux"
	c "github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/lucdoe/open-gateway/gateway/internal/proxy"
	srv "github.com/lucdoe/open-gateway/gateway/internal/server"
)

func main() {
	cfg, err := c.NewParser("./cmd/gateway/config.yaml").Parse()
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	router := mux.NewRouter()
	proxyService := proxy.NewProxyService()
	middlewares, err := srv.InitMiddleware()
	if err != nil {
		log.Fatalf("Failed to initialize middlewares: %v", err)
	}

	s := srv.NewServer(cfg, router, proxyService, middlewares)

	if err := s.Run(); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}
