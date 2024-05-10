package main

import (
	"log"

	c "github.com/lucdoe/open-gateway/gateway/internal/config"
	lp "github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
	srv "github.com/lucdoe/open-gateway/gateway/internal/server"
)

func main() {
	s := srv.NewServer()

	l := lp.NewLogger("server.log", nil)
	if err := l.Init(); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	s.Plugins.RegisterPlugin("logger", l)

	cfg, err := c.NewParser("./cmd/gateway/config.yaml").Parse()
	if err != nil {
		log.Fatalf("failed to parse configuration: %v", err)
	}

	s.SetupRoutes(cfg)

	if err := s.Run(); err != nil {
		log.Fatalf("failed to run the server: %v", err)
	}
}
