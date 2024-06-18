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

func TestSetupRoutes(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"starwars": {
				URL:      "https://swapi.dev/api",
				Protocol: "HTTPS",
				Plugins:  []string{"rate-limit", "logger", "cache", "cors"},
				Endpoints: []config.Endpoint{
					{
						Name:       "List People",
						HTTPMethod: "GET",
						Path:       "/people",
						QueryParams: []config.QueryParam{
							{Key: "page", Value: "1"},
						},
					},
					{
						Name:       "Get Person",
						HTTPMethod: "GET",
						Path:       "/people/{id}",
					},
					{
						Name:       "List Starships",
						HTTPMethod: "GET",
						Path:       "/starships",
						QueryParams: []config.QueryParam{
							{Key: "page", Value: "1"},
						},
					},
					{
						Name:       "Get Starship",
						HTTPMethod: "GET",
						Path:       "/starships/{id}",
					},
				},
			},
		},
	}

	router := mux.NewRouter()
	proxyService := &mockProxyService{}
	middlewares := map[string]Middleware{}

	s := NewServer(cfg, router, proxyService, middlewares)
	s.SetupRoutes(cfg)

	ts := httptest.NewServer(router)
	defer ts.Close()

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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, ts.URL+tc.path, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expected {
				t.Errorf("Expected status %d, got %d", tc.expected, resp.StatusCode)
			}
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

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"Item Present", []string{"apple", "banana", "orange"}, "banana", true},
		{"Item Absent", []string{"apple", "banana", "orange"}, "grape", false},
		{"Empty Slice", []string{}, "apple", false},
		{"Nil Slice", nil, "apple", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := contains(test.slice, test.item)
			assert.Equal(t, test.expected, result)
		})
	}
}
