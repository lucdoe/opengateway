package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
)

type ProxyService interface {
	ReverseProxy(targetURL string, w http.ResponseWriter, r *http.Request) error
}

type Server struct {
	Router      *mux.Router
	Middlewares map[string]Middleware
	Proxy       ProxyService
}

func NewServer(cfg *config.Config, router *mux.Router, proxy ProxyService, mws map[string]Middleware) *Server {
	server := &Server{
		Router:      router,
		Middlewares: mws,
		Proxy:       proxy,
	}
	server.SetupRoutes(cfg)
	return server
}

func (s *Server) SetupRoutes(cfg *config.Config) {
	for _, service := range cfg.Services {
		applyServiceMiddlewares(s, service.Plugins)

		for _, endpoint := range service.Endpoints {
			handler := MakeHandler(s.Proxy, service.URL)
			s.Router.HandleFunc(endpoint.Path, handler).Methods(endpoint.HTTPMethod)
		}
	}
}

func MakeHandler(proxy ProxyService, targetURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := proxy.ReverseProxy(targetURL, w, r); err != nil {
			http.Error(w, "Proxy error", http.StatusInternalServerError)
		}
	}
}

func applyServiceMiddlewares(s *Server, plugins []string) {
	orderedMiddlewareKeys := []string{"logger", "cors", "rate-limit", "cache"}

	for _, key := range orderedMiddlewareKeys {
		if _, ok := s.Middlewares[key]; ok && contains(plugins, key) {
			s.Router.Use(s.Middlewares[key].Middleware)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return true
		}
	}
	return false
}

func (s *Server) Run() error {
	log.Println("Starting server on http://localhost:4000")
	return http.ListenAndServe(":4000", s.Router)
}
