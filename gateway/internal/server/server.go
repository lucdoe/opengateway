package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
	plugin "github.com/lucdoe/open-gateway/gateway/internal/plugins"
	"github.com/lucdoe/open-gateway/gateway/internal/proxy"
)

type Server struct {
	Router  *mux.Router
	Plugins *plugin.Manager
}

func NewServer() *Server {
	return &Server{
		Router:  mux.NewRouter(),
		Plugins: plugin.NewManager(),
	}
}

func (s *Server) SetupRoutes(cfg *config.Config) {
	for _, service := range cfg.Services {
		serviceHandler := func(next http.Handler) http.Handler {
			return s.Plugins.ApplyPlugins(service.Plugins, next)
		}

		for _, endpoint := range service.Endpoints {
			targetURL := service.URL
			handler := func(w http.ResponseWriter, r *http.Request) {
				proxyService := proxy.NewProxyService()
				err := proxyService.ReverseProxy(targetURL, w, r)
				if err != nil {
					http.Error(w, "Proxy error", http.StatusInternalServerError)
				}
			}

			endpointHandler := serviceHandler(http.HandlerFunc(handler))

			wrappedHandler := s.Plugins.ApplyPlugins(endpoint.Plugins, endpointHandler)
			s.Router.HandleFunc(endpoint.Path, wrappedHandler.ServeHTTP).Methods(endpoint.HTTPMethod)
		}
	}
}

func (s *Server) Run() error {
	log.Println("Starting server on http://localhost:4000")
	return http.ListenAndServe(":4000", s.Router)
}
