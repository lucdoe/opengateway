package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/lucdoe/open-gateway/gateway/internal/proxy"
)

type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	return &Server{
		Router: mux.NewRouter(),
	}
}

func (s *Server) SetupRoutes(cfg *config.Config) {
	for serviceName, service := range cfg.Services {
		for _, endpoint := range service.Endpoints {
			targetURL := service.URL

			handler := func(w http.ResponseWriter, r *http.Request) {
				proxyService := proxy.NewProxyService()
				err := proxyService.ReverseProxy(targetURL, w, r)
				if err != nil {
					http.Error(w, "Proxy error", http.StatusInternalServerError)
					log.Printf("Error handling request to %s: %v", targetURL, err)
				}
			}

			s.Router.HandleFunc(endpoint.Path, handler).Methods(endpoint.HTTPMethod)
			log.Printf("Registered %s ->  %s", endpoint.HTTPMethod, targetURL+endpoint.Path)
		}
		log.Printf("Service %s setup completed", serviceName)
	}
}

func (s *Server) Run() error {
	log.Println("Starting server on http://localhost:4000")
	return http.ListenAndServe(":4000", s.Router)
}
