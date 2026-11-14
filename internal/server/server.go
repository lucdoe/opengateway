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
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"slices"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/breaker"
	"github.com/lucdoe/opengateway/internal/config"
)

// Middleware wraps a handler with one plugin's behaviour.
type Middleware func(next http.Handler) http.Handler

// ProxyFactory builds the handler that forwards requests to the upstream at
// target and gives up when the upstream has not started answering within
// timeout. It is called once per service when routes are set up.
type ProxyFactory func(target string, timeout time.Duration) (http.Handler, error)

type Server struct {
	// Addr is the address Run listens on, for example ":4000".
	Addr string
	// Settings holds the timeouts of the HTTP server and its shutdown.
	Settings    config.ServerConfig
	Router      *mux.Router
	Middlewares map[string]Middleware
	NewProxy    ProxyFactory
	// Closers are closed when Serve returns, after the requests in flight
	// have finished, for example to write out a buffered log.
	Closers []io.Closer
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
		Addr:        cfg.ListenAddr(),
		Settings:    cfg.Server,
		Router:      router,
		Middlewares: mws,
		NewProxy:    newProxy,
	}
	if err := server.SetupRoutes(cfg); err != nil {
		return nil, err
	}
	return server, nil
}

// SetupRoutes registers one route per configured endpoint. Each service gets
// one proxy to its upstream behind one circuit breaker, both shared by its
// endpoints, and each route gets its own plugin chain, so a plugin only runs
// for the services and endpoints that ask for it.
func (s *Server) SetupRoutes(cfg *config.Config) error {
	names := make([]string, 0, len(cfg.Services))
	for name := range cfg.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		service := cfg.Services[name]

		upstream, err := s.NewProxy(service.URL, cfg.Server.UpstreamTimeout)
		if err != nil {
			return fmt.Errorf("service %q: %w", name, err)
		}
		upstream = breaker.Wrap(name, cfg.Server.CircuitBreaker, upstream)

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

// Run listens on Addr and serves requests until ctx is cancelled. It then
// shuts down gracefully, see Serve.
func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(ctx, listener)
}

// Serve serves requests on listener until ctx is cancelled. It then stops
// accepting new connections and waits for the requests in flight to finish.
// If they have not finished within Settings.ShutdownTimeout, the remaining
// connections are closed and an error is returned.
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	defer s.close()
	httpServer := &http.Server{
		Handler:           s.Router,
		ReadHeaderTimeout: s.Settings.ReadHeaderTimeout,
		IdleTimeout:       s.Settings.IdleTimeout,
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.Serve(listener) }()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.Settings.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		httpServer.Close()
		return fmt.Errorf("requests still in flight after %s were cut off: %w", s.Settings.ShutdownTimeout, err)
	}
	return nil
}

func (s *Server) close() {
	for _, closer := range s.Closers {
		if err := closer.Close(); err != nil {
			log.Printf("closing on shutdown: %v", err)
		}
	}
}
