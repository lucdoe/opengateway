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
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// okProxy stands in for a real upstream: it answers every request with 200.
func okProxy(target string, timeout time.Duration) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil
}

// recordingMiddleware appends its name to calls every time it runs. When
// reject is set it answers 401 instead of calling the next handler.
type recordingMiddleware struct {
	name   string
	calls  *[]string
	reject bool
}

func (m *recordingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*m.calls = append(*m.calls, m.name)
		if m.reject {
			http.Error(w, "rejected", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func recordingMiddlewares(calls *[]string) map[string]Middleware {
	mws := map[string]Middleware{}
	for _, name := range pluginOrder {
		mws[name] = (&recordingMiddleware{name: name, calls: calls}).Middleware
	}
	return mws
}

func serve(s *Server, method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func TestSetupRoutes(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"starwars": {
				URL:     "https://swapi.dev/api",
				Plugins: []string{"rate-limit", "logger", "cache", "cors"},
				Endpoints: []config.Endpoint{
					{Name: "List People", HTTPMethod: "GET", Path: "/people"},
					{Name: "Get Person", HTTPMethod: "GET", Path: "/people/{id}"},
					{Name: "List Starships", HTTPMethod: "GET", Path: "/starships"},
					{Name: "Get Starship", HTTPMethod: "GET", Path: "/starships/{id}"},
				},
			},
		},
	}

	var calls []string
	s, err := NewServer(cfg, mux.NewRouter(), okProxy, recordingMiddlewares(&calls), nil)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"List People", "GET", "/people", http.StatusOK},
		{"Get Person", "GET", "/people/1", http.StatusOK},
		{"List Starships", "GET", "/starships", http.StatusOK},
		{"Get Starship", "GET", "/starships/1", http.StatusOK},
		{"Wrong method", "POST", "/people", http.StatusMethodNotAllowed},
		{"Unknown path", "GET", "/planets", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, serve(s, tc.method, tc.path).Code)
		})
	}
}

func TestPluginsApplyOnlyToTheirOwnService(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"full": {
				Subpath: "/full",
				Plugins: []string{"cache", "rate-limit", "cors", "logger"},
				Endpoints: []config.Endpoint{
					{Name: "Items", HTTPMethod: "GET", Path: "/items"},
				},
			},
			"plain": {
				Subpath: "/plain",
				Plugins: []string{"logger"},
				Endpoints: []config.Endpoint{
					{Name: "Items", HTTPMethod: "GET", Path: "/items"},
				},
			},
			"bare": {
				Subpath: "/bare",
				Endpoints: []config.Endpoint{
					{Name: "Items", HTTPMethod: "GET", Path: "/items"},
				},
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{"all plugins, in pipeline order", "/full/items", []string{"logger", "cors", "rate-limit", "cache"}},
		{"only its own plugin, once", "/plain/items", []string{"logger"}},
		{"no plugins", "/bare/items", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			s, err := NewServer(cfg, mux.NewRouter(), okProxy, recordingMiddlewares(&calls), nil)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, serve(s, "GET", tc.path).Code)
			assert.Equal(t, tc.expected, calls)
		})
	}
}

func TestEndpointPluginsAndAuth(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api": {
				Plugins: []string{"logger"},
				Endpoints: []config.Endpoint{
					{Name: "Open", HTTPMethod: "GET", Path: "/open"},
					{Name: "Cached", HTTPMethod: "GET", Path: "/cached", Plugins: []string{"cache", "logger"}},
					{Name: "Private", HTTPMethod: "GET", Path: "/private", Plugins: []string{"cache"}, Auth: config.AuthConfig{ApplyAuth: true}},
				},
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{"service plugins only", "/open", []string{"logger"}},
		{"endpoint plugins are added without duplicates", "/cached", []string{"logger", "cache"}},
		{"ApplyAuth adds auth ahead of cache", "/private", []string{"logger", "auth", "cache"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			s, err := NewServer(cfg, mux.NewRouter(), okProxy, recordingMiddlewares(&calls), nil)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, serve(s, "GET", tc.path).Code)
			assert.Equal(t, tc.expected, calls)
		})
	}

	t.Run("a rejecting auth plugin stops the request", func(t *testing.T) {
		var calls []string
		mws := recordingMiddlewares(&calls)
		mws["auth"] = (&recordingMiddleware{name: "auth", calls: &calls, reject: true}).Middleware
		s, err := NewServer(cfg, mux.NewRouter(), okProxy, mws, nil)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, serve(s, "GET", "/private").Code)
		assert.Equal(t, []string{"logger", "auth"}, calls)

		calls = nil
		assert.Equal(t, http.StatusOK, serve(s, "GET", "/open").Code)
		assert.Equal(t, []string{"logger"}, calls)
	})
}

// recordingScope stands in for auth.RequireScope and records "scope:<scope>".
func recordingScope(calls *[]string) ScopeCheck {
	return func(scope string) Middleware {
		return (&recordingMiddleware{name: "scope:" + scope, calls: calls}).Middleware
	}
}

func TestEndpointScopeRunsRightAfterAuth(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api": {
				Plugins: []string{"logger", "cache"},
				Endpoints: []config.Endpoint{
					{Name: "Admin", HTTPMethod: "GET", Path: "/admin", Auth: config.AuthConfig{ApplyAuth: true, Scope: "admin"}},
					{Name: "Private", HTTPMethod: "GET", Path: "/private", Auth: config.AuthConfig{ApplyAuth: true}},
				},
			},
		},
	}
	var calls []string
	s, err := NewServer(cfg, mux.NewRouter(), okProxy, recordingMiddlewares(&calls), recordingScope(&calls))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, serve(s, "GET", "/admin").Code)
	assert.Equal(t, []string{"logger", "auth", "scope:admin", "cache"}, calls)

	calls = nil
	serve(s, "GET", "/private")
	assert.Equal(t, []string{"logger", "auth", "cache"}, calls, "no scope, no check")

	_, err = NewServer(cfg, mux.NewRouter(), okProxy, recordingMiddlewares(new([]string)), nil)
	assert.ErrorContains(t, err, `service "api", endpoint "Admin": Auth.Scope is set but no scope check is registered`)
}

func TestSetupRoutesRejectsUnknownPlugins(t *testing.T) {
	endpoint := config.Endpoint{Name: "Items", HTTPMethod: "GET", Path: "/items"}

	tests := []struct {
		name        string
		service     config.Service
		middlewares map[string]Middleware
	}{
		{
			name:        "misspelled service plugin",
			service:     config.Service{Plugins: []string{"loger"}, Endpoints: []config.Endpoint{endpoint}},
			middlewares: recordingMiddlewares(new([]string)),
		},
		{
			name: "misspelled endpoint plugin",
			service: config.Service{Endpoints: []config.Endpoint{
				{Name: "Items", HTTPMethod: "GET", Path: "/items", Plugins: []string{"cash"}},
			}},
			middlewares: recordingMiddlewares(new([]string)),
		},
		{
			name:        "plugin that was never registered",
			service:     config.Service{Plugins: []string{"logger"}, Endpoints: []config.Endpoint{endpoint}},
			middlewares: map[string]Middleware{},
		},
		{
			name: "ApplyAuth without an auth plugin",
			service: config.Service{Endpoints: []config.Endpoint{
				{Name: "Items", HTTPMethod: "GET", Path: "/items", Auth: config.AuthConfig{ApplyAuth: true}},
			}},
			middlewares: map[string]Middleware{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{Services: map[string]config.Service{"api": tc.service}}

			s, err := NewServer(cfg, mux.NewRouter(), okProxy, tc.middlewares, nil)

			assert.Nil(t, s)
			assert.ErrorContains(t, err, "unknown plugin")
			assert.ErrorContains(t, err, `service "api", endpoint "Items"`)
		})
	}
}

func TestNewServerListensOnTheConfiguredPort(t *testing.T) {
	cfg := &config.Config{Server: config.ServerConfig{Port: 8443}}

	s, err := NewServer(cfg, mux.NewRouter(), okProxy, map[string]Middleware{}, nil)

	assert.NoError(t, err)
	assert.Equal(t, ":8443", s.Addr)
}

func TestSetupRoutesBuildsOneProxyPerService(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"starwars": {
				URL:     "https://swapi.dev",
				Subpath: "/api",
				Endpoints: []config.Endpoint{
					{Name: "List People", HTTPMethod: "GET", Path: "/people"},
					{Name: "Get Person", HTTPMethod: "GET", Path: "/people/{id}"},
				},
			},
			"other": {
				URL:     "http://localhost:3001",
				Subpath: "/api",
				Endpoints: []config.Endpoint{
					{Name: "List Worlds", HTTPMethod: "GET", Path: "/worlds"},
				},
			},
		},
	}

	var targets []string
	served := map[string]string{}
	newProxy := func(target string, timeout time.Duration) (http.Handler, error) {
		targets = append(targets, target)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			served[r.URL.Path] = target
		}), nil
	}

	s, err := NewServer(cfg, mux.NewRouter(), newProxy, map[string]Middleware{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"http://localhost:3001", "https://swapi.dev"}, targets)

	for _, path := range []string{"/api/people", "/api/people/7", "/api/worlds"} {
		assert.Equal(t, http.StatusOK, serve(s, "GET", path).Code)
	}
	assert.Equal(t, map[string]string{
		"/api/people":   "https://swapi.dev",
		"/api/people/7": "https://swapi.dev",
		"/api/worlds":   "http://localhost:3001",
	}, served)
}

func TestSetupRoutesFailsWhenAProxyCannotBeBuilt(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"broken": {URL: "not a url", Endpoints: []config.Endpoint{{Name: "Items", HTTPMethod: "GET", Path: "/items"}}},
		},
	}
	newProxy := func(target string, timeout time.Duration) (http.Handler, error) {
		return nil, errors.New("invalid upstream URL")
	}

	s, err := NewServer(cfg, mux.NewRouter(), newProxy, map[string]Middleware{}, nil)

	assert.Nil(t, s)
	assert.ErrorContains(t, err, `service "broken": invalid upstream URL`)
}

// The route pattern contains {id}; the upstream must receive the real value.
func TestRoutesForwardPathParametersToTheUpstream(t *testing.T) {
	var upstreamPaths []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamPaths = append(upstreamPaths, r.URL.RequestURI())
		w.Write([]byte("from upstream"))
	}))
	defer upstream.Close()

	cfg := &config.Config{
		Services: map[string]config.Service{
			"starwars": {
				URL:     upstream.URL,
				Subpath: "/api",
				Endpoints: []config.Endpoint{
					{Name: "List People", HTTPMethod: "GET", Path: "/people"},
					{Name: "Get Person", HTTPMethod: "GET", Path: "/people/{id}"},
				},
			},
		},
	}

	s, err := NewServer(cfg, mux.NewRouter(), proxy.NewHandler, map[string]Middleware{}, nil)
	assert.NoError(t, err)

	for _, path := range []string{"/api/people/7", "/api/people?page=2"} {
		w := serve(s, "GET", path)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "from upstream", w.Body.String())
	}
	assert.Equal(t, []string{"/api/people/7", "/api/people?page=2"}, upstreamPaths)
}

func TestNewServerPassesTheUpstreamTimeoutToEveryProxy(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{UpstreamTimeout: 7 * time.Second},
		Services: map[string]config.Service{
			"a": {URL: "http://a.test", Endpoints: []config.Endpoint{{Name: "A", HTTPMethod: "GET", Path: "/a"}}},
			"b": {URL: "http://b.test", Endpoints: []config.Endpoint{{Name: "B", HTTPMethod: "GET", Path: "/b"}}},
		},
	}

	var timeouts []time.Duration
	newProxy := func(target string, timeout time.Duration) (http.Handler, error) {
		timeouts = append(timeouts, timeout)
		return http.NotFoundHandler(), nil
	}

	_, err := NewServer(cfg, mux.NewRouter(), newProxy, map[string]Middleware{}, nil)

	assert.NoError(t, err)
	assert.Equal(t, []time.Duration{7 * time.Second, 7 * time.Second}, timeouts)
}

func TestEachServiceHasItsOwnCircuitBreaker(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			CircuitBreaker: config.CircuitBreakerConfig{ConsecutiveFailures: 2, OpenTimeout: time.Minute, HalfOpenRequests: 1},
		},
		Services: map[string]config.Service{
			"down": {URL: "http://down.test", Endpoints: []config.Endpoint{
				{Name: "A", HTTPMethod: "GET", Path: "/down/a"},
				{Name: "B", HTTPMethod: "GET", Path: "/down/b"},
			}},
			"up": {URL: "http://up.test", Endpoints: []config.Endpoint{{Name: "C", HTTPMethod: "GET", Path: "/up"}}},
		},
	}
	newProxy := func(target string, timeout time.Duration) (http.Handler, error) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if target == "http://down.test" {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			w.WriteHeader(http.StatusOK)
		}), nil
	}
	var calls []string
	s, err := NewServer(cfg, mux.NewRouter(), newProxy, recordingMiddlewares(&calls), nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadGateway, serve(s, "GET", "/down/a").Code)
	assert.Equal(t, http.StatusBadGateway, serve(s, "GET", "/down/b").Code)

	assert.Equal(t, http.StatusServiceUnavailable, serve(s, "GET", "/down/a").Code, "failures on both endpoints opened the service's breaker")
	assert.Equal(t, http.StatusServiceUnavailable, serve(s, "GET", "/down/b").Code)
	assert.Equal(t, http.StatusOK, serve(s, "GET", "/up").Code, "another service is not affected")
}

// runningServer serves a single route, GET /slow, whose handler reports on
// started and then blocks until release is closed.
type runningServer struct {
	url     string
	started chan struct{}
	release chan struct{}
	stop    context.CancelFunc
	result  chan error
}

func startServer(t *testing.T, settings config.ServerConfig, closers ...io.Closer) *runningServer {
	t.Helper()
	rs := &runningServer{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
		result:  make(chan error, 1),
	}

	router := mux.NewRouter()
	router.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		rs.started <- struct{}{}
		<-rs.release
		w.Write([]byte("finished"))
	})
	router.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	rs.url = "http://" + listener.Addr().String()

	ctx, cancel := context.WithCancel(context.Background())
	rs.stop = cancel
	t.Cleanup(cancel)

	s := &Server{Settings: settings, Router: router, Closers: closers}
	go func() { rs.result <- s.Serve(ctx, listener) }()
	return rs
}

func (rs *runningServer) waitForResult(t *testing.T) error {
	t.Helper()
	select {
	case err := <-rs.result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("the server did not stop")
		return nil
	}
}

// closeRecorder notes when it was closed and whether the slow request had
// finished by then.
type closeRecorder struct {
	finished *atomic.Bool
	closed   atomic.Bool
	err      error
	sawDone  atomic.Bool
}

func (c *closeRecorder) Close() error {
	c.sawDone.Store(c.finished.Load())
	c.closed.Store(true)
	return c.err
}

// A buffered log must be written out after the last request, not before.
func TestServeClosesTheClosersAfterRequestsInFlightFinish(t *testing.T) {
	var finished atomic.Bool
	first := &closeRecorder{finished: &finished}
	failing := &closeRecorder{finished: &finished, err: errors.New("disk full")}
	rs := startServer(t, config.ServerConfig{ShutdownTimeout: 5 * time.Second}, failing, first)

	done := make(chan struct{})
	go func() {
		defer close(done)
		res, err := http.Get(rs.url + "/slow")
		if err == nil {
			res.Body.Close()
		}
	}()
	<-rs.started
	rs.stop()
	time.Sleep(50 * time.Millisecond)
	assert.False(t, first.closed.Load(), "nothing is closed while a request is in flight")

	finished.Store(true)
	close(rs.release)
	<-done

	assert.NoError(t, rs.waitForResult(t))
	assert.True(t, first.closed.Load())
	assert.True(t, first.sawDone.Load(), "closed after the request finished")
	assert.True(t, failing.closed.Load(), "a closer that fails does not stop the others")
}

func TestServeAnswersRequestsUntilStopped(t *testing.T) {
	rs := startServer(t, config.ServerConfig{ShutdownTimeout: time.Second})

	res, err := http.Get(rs.url + "/fast")
	require.NoError(t, err)
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	assert.Equal(t, "ok", string(body))

	rs.stop()

	assert.NoError(t, rs.waitForResult(t))
	_, err = http.Get(rs.url + "/fast")
	assert.Error(t, err, "the server must not accept requests after it stopped")
}

func TestServeLetsRequestsInFlightFinishOnShutdown(t *testing.T) {
	rs := startServer(t, config.ServerConfig{ShutdownTimeout: 5 * time.Second})

	type response struct {
		body string
		err  error
	}
	responses := make(chan response, 1)
	go func() {
		res, err := http.Get(rs.url + "/slow")
		if err != nil {
			responses <- response{err: err}
			return
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		responses <- response{body: string(body), err: err}
	}()

	<-rs.started
	rs.stop()

	select {
	case err := <-rs.result:
		t.Fatalf("the server stopped while a request was in flight: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(rs.release)

	got := <-responses
	assert.NoError(t, got.err)
	assert.Equal(t, "finished", got.body)
	assert.NoError(t, rs.waitForResult(t))
}

func TestServeCutsOffRequestsAfterTheShutdownTimeout(t *testing.T) {
	rs := startServer(t, config.ServerConfig{ShutdownTimeout: 50 * time.Millisecond})
	defer close(rs.release)

	go func() {
		if res, err := http.Get(rs.url + "/slow"); err == nil {
			res.Body.Close()
		}
	}()

	<-rs.started
	rs.stop()

	err := rs.waitForResult(t)
	assert.ErrorContains(t, err, "requests still in flight after 50ms were cut off")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestServeDropsClientsThatDoNotSendTheirHeaders(t *testing.T) {
	rs := startServer(t, config.ServerConfig{ReadHeaderTimeout: 50 * time.Millisecond, ShutdownTimeout: time.Second})

	conn, err := net.Dial("tcp", rs.url[len("http://"):])
	require.NoError(t, err)
	defer conn.Close()

	// Start a request but never finish its headers.
	_, err = conn.Write([]byte("GET /fast HTTP/1.1\r\nHost: gateway.test\r\n"))
	require.NoError(t, err)

	// The server must end the connection on its own; reading reaches the end
	// of the stream well before this deadline.
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, err = io.ReadAll(conn)
	assert.NoError(t, err, "the connection was still open after the header timeout")
}

func TestRunFailsWhenTheAddressIsTaken(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	s := &Server{Addr: listener.Addr().String(), Router: mux.NewRouter()}

	assert.ErrorContains(t, s.Run(context.Background()), "address already in use")
}
