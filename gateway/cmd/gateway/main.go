package main

import (
	"log"

	c "github.com/lucdoe/open-gateway/gateway/internal/config"
	srv "github.com/lucdoe/open-gateway/gateway/internal/server"
)

func main() {
	cfg, err := c.NewParser("./cmd/gateway/config.yaml").Parse()
	if err != nil {
		log.Fatalf("failed to parse configuration: %v", err)
	}

	s, err := srv.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	if err := s.Run(); err != nil {
		log.Fatalf("failed to run the server: %v", err)
	}
}
