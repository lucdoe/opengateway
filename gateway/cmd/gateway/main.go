package main

import (
	"log"

	c "github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/lucdoe/open-gateway/gateway/internal/server"
)

func main() {
	p := c.NewParser("./cmd/gateway/config.yaml")

	cfg, err := p.Parse()
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	server := server.NewServer()

	server.SetupRoutes(cfg)

	err = server.Run()
	if err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}
