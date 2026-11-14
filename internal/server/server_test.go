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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProxyService struct {
	err error
}

func (m *mockProxyService) ReverseProxy(targetURL string, w http.ResponseWriter, r *http.Request) error {
	return m.err
}

type MockServerRunner struct {
	mock.Mock
}

func (m *MockServerRunner) ListenAndServe(addr string, handler http.Handler) error {
	args := m.Called(addr, handler)
	return args.Error(0)
}

func TestMakeHandler(t *testing.T) {
	tests := []struct {
		name         string
		proxyError   error
		expectedCode int
	}{
		{"success", nil, http.StatusOK},
		{"proxy error", http.ErrHandlerTimeout, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			w := httptest.NewRecorder()
			proxy := &mockProxyService{err: tt.proxyError}
			handler := MakeHandler(proxy, "http://target.url")

			handler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status %d; got %d", tt.expectedCode, resp.StatusCode)
			}
		})
	}
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
		mws[name] = &recordingMiddleware{name: name, calls: calls}
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
				URL:      "https://swapi.dev/api",
				Protocol: "HTTPS",
				Plugins:  []string{"rate-limit", "logger", "cache", "cors"},
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
	s, err := NewServer(cfg, mux.NewRouter(), &mockProxyService{}, recordingMiddlewares(&calls))
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
			s, err := NewServer(cfg, mux.NewRouter(), &mockProxyService{}, recordingMiddlewares(&calls))
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
			s, err := NewServer(cfg, mux.NewRouter(), &mockProxyService{}, recordingMiddlewares(&calls))
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, serve(s, "GET", tc.path).Code)
			assert.Equal(t, tc.expected, calls)
		})
	}

	t.Run("a rejecting auth plugin stops the request", func(t *testing.T) {
		var calls []string
		mws := recordingMiddlewares(&calls)
		mws["auth"] = &recordingMiddleware{name: "auth", calls: &calls, reject: true}
		s, err := NewServer(cfg, mux.NewRouter(), &mockProxyService{}, mws)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, serve(s, "GET", "/private").Code)
		assert.Equal(t, []string{"logger", "auth"}, calls)

		calls = nil
		assert.Equal(t, http.StatusOK, serve(s, "GET", "/open").Code)
		assert.Equal(t, []string{"logger"}, calls)
	})
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

			s, err := NewServer(cfg, mux.NewRouter(), &mockProxyService{}, tc.middlewares)

			assert.Nil(t, s)
			assert.ErrorContains(t, err, "unknown plugin")
			assert.ErrorContains(t, err, `service "api", endpoint "Items"`)
		})
	}
}

func TestServerRun(t *testing.T) {
	mockRunner := new(MockServerRunner)
	mockRunner.On("ListenAndServe", ":4000", mock.Anything).Return(nil)

	s := Server{
		Router: mux.NewRouter(),
		Runner: mockRunner,
	}

	err := s.Run()
	assert.NoError(t, err)
	mockRunner.AssertExpectations(t)
}
