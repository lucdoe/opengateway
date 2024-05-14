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

package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockConfigLoader struct {
	mock.Mock
}

func (m *MockConfigLoader) LoadConfig(path string) (*config.Config, error) {
	args := m.Called(path)
	return args.Get(0).(*config.Config), args.Error(1)
}

type MockRouter struct {
	*mux.Router
}

func NewMockRouter() *MockRouter {
	return &MockRouter{mux.NewRouter()}
}

type MockProxyService struct {
	mock.Mock
}

func (m *MockProxyService) ReverseProxy(serviceName string, rw http.ResponseWriter, req *http.Request) error {
	args := m.Called(serviceName, rw, req)
	return args.Error(0)
}

type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Increment(key string, window time.Duration) (int64, error) {
	args := m.Called(key, window)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Set(key string, value string, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) GenerateCacheKey(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

func TestInitializeServerSuccess(t *testing.T) {
	mockConfigLoader := new(MockConfigLoader)
	mockRouter := NewMockRouter()
	mockProxyService := new(MockProxyService)
	mockProxyService.On("ReverseProxy", mock.Anything, mock.Anything).Return(nil)

	mockCacheService := new(MockCacheService)
	mockCacheService.On("Increment", mock.Anything, mock.Anything).Return(1, nil)

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
							{
								Key:   "page",
								Value: "1",
							},
						},
						Auth: config.AuthConfig{ApplyAuth: false},
					},
				},
			},
		},
		Plugins: config.PluginConfig{
			LoggerConfig: config.LoggerConfig{
				FilePath:  "server_test.go",
				ErrOutput: "stderr",
			},
			JWTConfig: config.JWTConfig{
				SecretKey:     "your-256-bit-secret",
				SigningMethod: "HS256",
				Issuer:        "ExampleIssuer",
				Audience:      "ExampleAudience",
				Scope:         "ExampleScope",
			},
			RateLimitConfig: config.RateLimitConfig{
				Limit:  100,
				Window: 60 * time.Second,
			},
			CORSConfig: config.CORSConfig{
				Origins: "*",
				Methods: "GET, POST, PUT, DELETE, OPTIONS",
				Headers: "Content-Type, Authorization",
			},
		},
	}

	mockConfigLoader.On("LoadConfig", "./cmd/gateway/config.yaml").Return(cfg, nil)

	deps := ServerDependencies{
		ConfigLoader: mockConfigLoader,
		Router:       mockRouter.Router,
		ProxyService: mockProxyService,
		CacheService: mockCacheService,
	}

	server, err := InitializeServer(deps)
	assert.NoError(t, err)
	assert.NotNil(t, server)
}
