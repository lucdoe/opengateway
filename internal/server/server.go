// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/breaker"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/httpx"
	"github.com/lucdoe/opengateway/internal/requestid"
)

// Middleware wraps a handler with one plugin's behaviour.
type Middleware func(next http.Handler) http.Handler

// ProxyFactory builds the handler that forwards requests to the upstream at
// target and gives up when the upstream has not started answering within
// timeout. It is called once per service when routes are set up.
type ProxyFactory func(target string, timeout time.Duration) (http.Handler, error)

// ScopeCheck builds the check for an endpoint's Auth.Scope. It runs right
// after the auth plugin, whose accepted token it inspects.
type ScopeCheck func(scope string) Middleware

type Server struct {
	// Addr is the address Run listens on, for example ":4000".
	Addr string
	// Settings holds the timeouts of the HTTP server and its shutdown.
	Settings    config.ServerConfig
	Router      *mux.Router
	Middlewares map[string]Middleware
	NewProxy    ProxyFactory
	ScopeCheck  ScopeCheck
	// Closers are closed when Serve returns, after the requests in flight
	// have finished, for example to write out a buffered log.
	Closers []io.Closer
}

// Handler is what Serve answers requests with: Router behind the request
// ID, so every response carries one, including a 404.
func (s *Server) Handler() http.Handler {
	return requestid.Middleware(s.Router)
}

// HealthPath answers 200 while the gateway is running, for load balancers
// and container health checks. It runs no plugins and calls no upstream.
const HealthPath = config.ReservedPathPrefix + "health"

func NewServer(cfg *config.Config, router *mux.Router, newProxy ProxyFactory, mws map[string]Middleware, scopeCheck ScopeCheck) (*Server, error) {
	server := &Server{
		Addr:        cfg.ListenAddr(),
		Settings:    cfg.Server,
		Router:      router,
		Middlewares: mws,
		NewProxy:    newProxy,
		ScopeCheck:  scopeCheck,
	}
	if err := server.SetupRoutes(cfg); err != nil {
		return nil, err
	}
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusNotFound, "Not found")
	})
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
	})
	return server, nil
}

// SetupRoutes registers one route per configured endpoint. Each service gets
// one proxy to its upstream behind one circuit breaker, both shared by its
// endpoints, and each route gets its own plugin chain, so a plugin only runs
// for the services and endpoints that ask for it.
//
// A path with a cors endpoint also gets an OPTIONS route with that endpoint's
// chain, so browser preflights reach the cors plugin, unless an endpoint
// declares OPTIONS for the path itself.
func (s *Server) SetupRoutes(cfg *config.Config) error {
	s.Router.HandleFunc(HealthPath, health).Methods(http.MethodGet, http.MethodHead)

	hasOptions := map[string]bool{}
	for _, service := range cfg.Services {
		for _, endpoint := range service.Endpoints {
			if strings.EqualFold(endpoint.HTTPMethod, http.MethodOptions) {
				hasOptions[service.Subpath+endpoint.Path] = true
			}
		}
	}

	order := cfg.PluginsInOrder()
	for _, name := range cfg.ServiceNames() {
		service := cfg.Services[name]

		upstream, err := s.NewProxy(service.URL, cfg.Server.UpstreamTimeout)
		if err != nil {
			return fmt.Errorf("service %q: %w", name, err)
		}
		upstream = breaker.Wrap(name, cfg.Server.CircuitBreaker, upstream)

		for _, endpoint := range service.Endpoints {
			plugins, err := s.endpointPlugins(order, service, endpoint)
			if err != nil {
				return fmt.Errorf("service %q, endpoint %q: %w", name, endpoint.Name, err)
			}
			var afterAuth Middleware
			if scope := endpoint.Auth.Scope; scope != "" {
				if s.ScopeCheck == nil {
					return fmt.Errorf("service %q, endpoint %q: Auth.Scope is set but no scope check is registered", name, endpoint.Name)
				}
				afterAuth = s.ScopeCheck(scope)
			}

			handler := s.chain(plugins, upstream, afterAuth)
			path := service.Subpath + endpoint.Path
			s.Router.Handle(path, handler).Methods(endpoint.HTTPMethod)
			if slices.Contains(plugins, config.PluginCORS) && !hasOptions[path] {
				s.Router.Handle(path, handler).Methods(http.MethodOptions)
				hasOptions[path] = true
			}
		}
	}
	return nil
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(`{"status":"ok"}` + "\n"))
}

// endpointPlugins returns the plugins enabled for an endpoint, in order:
// the service's plugins, the endpoint's own plugins, and auth when the
// endpoint sets ApplyAuth. A name with no registered plugin is an error.
func (s *Server) endpointPlugins(order []string, service config.Service, endpoint config.Endpoint) ([]string, error) {
	requested := append(append([]string{}, service.Plugins...), endpoint.Plugins...)
	if endpoint.Auth.ApplyAuth {
		requested = append(requested, config.PluginAuth)
	}

	for _, name := range requested {
		if _, ok := s.Middlewares[name]; !ok {
			return nil, fmt.Errorf("unknown plugin %q", name)
		}
	}

	enabled := make([]string, 0, len(requested))
	for _, name := range order {
		if slices.Contains(requested, name) {
			enabled = append(enabled, name)
		}
	}
	return enabled, nil
}

// chain wraps handler in the named plugins so the first name runs first.
// afterAuth, if set, runs right after the auth plugin.
func (s *Server) chain(plugins []string, handler http.Handler, afterAuth Middleware) http.Handler {
	for i := len(plugins) - 1; i >= 0; i-- {
		if plugins[i] == config.PluginAuth && afterAuth != nil {
			handler = afterAuth(handler)
		}
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
		Handler:           s.Handler(),
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
