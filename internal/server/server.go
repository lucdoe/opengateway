// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type ServerRunner interface {
	ListenAndServe(addr string, handler http.Handler) error
}

type DefaultServerRunner struct{}

func (dsr *DefaultServerRunner) ListenAndServe(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}

type Server struct {
	Router      *mux.Router
	Middlewares map[string]Middleware
	Proxy       ProxyService
	Runner      ServerRunner
}

func NewServer(cfg *config.Config, router *mux.Router, proxy ProxyService, mws map[string]Middleware) *Server {
	server := &Server{
		Router:      router,
		Middlewares: mws,
		Proxy:       proxy,
		Runner:      &DefaultServerRunner{},
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
		if _, ok := s.Middlewares[key]; ok && Contains(plugins, key) {
			s.Router.Use(s.Middlewares[key].Middleware)
		}
	}
}

func Contains(slice []string, item string) bool {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return true
		}
	}
	return false
}

func (s *Server) Run() error {
	log.Println("Starting server on http://localhost:4000")
	return s.Runner.ListenAndServe(":4000", s.Router)
}
