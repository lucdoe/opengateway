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
	"fmt"
	"net/http"
	"slices"
	"sort"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
)

// Middleware wraps a handler with one plugin's behaviour.
type Middleware func(next http.Handler) http.Handler

// ProxyFactory builds the handler that forwards requests to the upstream at
// target. It is called once per service when routes are set up.
type ProxyFactory func(target string) (http.Handler, error)

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
	NewProxy    ProxyFactory
	Runner      ServerRunner
}

// pluginOrder is the order plugins run in for a request, outermost first.
// Auth runs before cache so a cached response is never served to a caller
// that has not been authenticated.
var pluginOrder = []string{
	config.PluginLogger,
	config.PluginCORS,
	config.PluginRateLimit,
	config.PluginAuth,
	config.PluginCache,
}

func NewServer(cfg *config.Config, router *mux.Router, newProxy ProxyFactory, mws map[string]Middleware) (*Server, error) {
	server := &Server{
		Router:      router,
		Middlewares: mws,
		NewProxy:    newProxy,
		Runner:      &DefaultServerRunner{},
	}
	if err := server.SetupRoutes(cfg); err != nil {
		return nil, err
	}
	return server, nil
}

// SetupRoutes registers one route per configured endpoint. Each service gets
// one proxy to its upstream, shared by its endpoints, and each route gets its
// own plugin chain, so a plugin only runs for the services and endpoints that
// ask for it.
func (s *Server) SetupRoutes(cfg *config.Config) error {
	names := make([]string, 0, len(cfg.Services))
	for name := range cfg.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		service := cfg.Services[name]

		upstream, err := s.NewProxy(service.URL)
		if err != nil {
			return fmt.Errorf("service %q: %w", name, err)
		}

		for _, endpoint := range service.Endpoints {
			plugins, err := s.endpointPlugins(service, endpoint)
			if err != nil {
				return fmt.Errorf("service %q, endpoint %q: %w", name, endpoint.Name, err)
			}

			handler := s.chain(plugins, upstream)
			s.Router.Handle(service.Subpath+endpoint.Path, handler).Methods(endpoint.HTTPMethod)
		}
	}
	return nil
}

// endpointPlugins returns the plugins enabled for an endpoint in the order
// they run: the service's plugins, the endpoint's own plugins, and auth when
// the endpoint sets ApplyAuth. A name with no registered plugin is an error.
func (s *Server) endpointPlugins(service config.Service, endpoint config.Endpoint) ([]string, error) {
	requested := append(append([]string{}, service.Plugins...), endpoint.Plugins...)
	if endpoint.Auth.ApplyAuth {
		requested = append(requested, config.PluginAuth)
	}

	for _, name := range requested {
		if _, ok := s.Middlewares[name]; !ok || !slices.Contains(pluginOrder, name) {
			return nil, fmt.Errorf("unknown plugin %q", name)
		}
	}

	enabled := make([]string, 0, len(requested))
	for _, name := range pluginOrder {
		if slices.Contains(requested, name) {
			enabled = append(enabled, name)
		}
	}
	return enabled, nil
}

// chain wraps handler in the named plugins so the first name runs first.
func (s *Server) chain(plugins []string, handler http.Handler) http.Handler {
	for i := len(plugins) - 1; i >= 0; i-- {
		handler = s.Middlewares[plugins[i]](handler)
	}
	return handler
}

func (s *Server) Run() error {
	return s.Runner.ListenAndServe(":4000", s.Router)
}
