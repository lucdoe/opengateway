package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/lucdoe/open-gateway/gateway/internal/proxy"
)

type Server struct {
	Router      *mux.Router
	Middlewares map[string]Middleware
}

func NewServer(cfg *config.Config) (*Server, error) {
	router := mux.NewRouter()
	mws, err := InitMiddleware()
	if err != nil {
		return nil, err
	}

	server := &Server{
		Router:      router,
		Middlewares: mws,
	}
	server.setupRoutes(cfg)
	return server, nil
}

func (s *Server) setupRoutes(cfg *config.Config) {
	proxyService := proxy.NewProxyService()

	for _, service := range cfg.Services {
		applyServiceMiddlewares(s, service.Plugins)

		for _, endpoint := range service.Endpoints {
			targetURL := service.URL
			handler := func(w http.ResponseWriter, r *http.Request) {
				err := proxyService.ReverseProxy(targetURL, w, r)
				if err != nil {
					http.Error(w, "Proxy error", http.StatusInternalServerError)
				}
			}

			s.Router.HandleFunc(endpoint.Path, handler).Methods(endpoint.HTTPMethod)
		}
	}
}

func applyServiceMiddlewares(s *Server, plugins []string) {
	orderedMiddlewareKeys := []string{"logger", "rate-limit", "cache"}

	includedMiddlewares := make(map[string]bool)
	for _, plugin := range plugins {
		includedMiddlewares[plugin] = true
	}

	for _, key := range orderedMiddlewareKeys {
		if includedMiddlewares[key] {
			middleware, exists := s.Middlewares[key]
			if !exists {
				log.Printf("Middleware %s not found", key)
				continue
			}
			s.Router.Use(middleware.Middleware)
		}
	}
}

func (s *Server) Run() error {
	log.Println("Starting server on http://localhost:4000")
	return http.ListenAndServe(":4000", s.Router)
}
