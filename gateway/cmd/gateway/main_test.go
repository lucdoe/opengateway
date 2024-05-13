package main

import (
	"errors"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lucdoe/open-gateway/gateway/internal/config"
	"github.com/lucdoe/open-gateway/gateway/internal/proxy"
	"github.com/lucdoe/open-gateway/gateway/internal/server"
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

func TestLoadConfig(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	expectedConfig := &config.Config{}
	mockLoader.On("LoadConfig", "./cmd/gateway/config.yaml").Return(expectedConfig, nil)

	cfg, err := mockLoader.LoadConfig("./cmd/gateway/config.yaml")
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, cfg)
}

type MockMiddlewareInitializer struct {
	mock.Mock
}

func (m *MockMiddlewareInitializer) InitMiddleware(cfg server.MiddlewareConfig) (map[string]server.Middleware, error) {
	args := m.Called(cfg)
	return args.Get(0).(map[string]server.Middleware), args.Error(1)
}

func TestInitMiddleware(t *testing.T) {
	mockInitializer := new(MockMiddlewareInitializer)
	expectedMiddlewares := map[string]server.Middleware{}
	mockInitializer.On("InitMiddleware", mock.Anything).Return(expectedMiddlewares, nil)

	middlewares, err := mockInitializer.InitMiddleware(server.MiddlewareConfig{})
	assert.NoError(t, err)
	assert.Equal(t, expectedMiddlewares, middlewares)
}

func TestInitializeServer(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	mockMiddlewareInit := new(MockMiddlewareInitializer)
	mockProxyService := proxy.NewProxyService()

	config := &config.Config{}
	middlewareMap := map[string]server.Middleware{}

	mockLoader.On("LoadConfig", "./cmd/gateway/config.yaml").Return(config, nil)
	mockMiddlewareInit.On("InitMiddleware", mock.Anything).Return(middlewareMap, nil)

	deps := ServerDependencies{
		ConfigLoader:          mockLoader,
		MiddlewareInitializer: mockMiddlewareInit,
		Router:                mux.NewRouter(),
		ProxyService:          mockProxyService,
	}

	server, err := InitializeServer(deps)
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestInitializeServerConfigLoadError(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	mockMiddlewareInit := new(MockMiddlewareInitializer)
	mockProxyService := proxy.NewProxyService()

	config := &config.Config{}
	mockLoader.On("LoadConfig", "./cmd/gateway/config.yaml").Return(config, errors.New("failed to load config"))

	deps := ServerDependencies{
		ConfigLoader:          mockLoader,
		MiddlewareInitializer: mockMiddlewareInit,
		Router:                mux.NewRouter(),
		ProxyService:          mockProxyService,
	}

	_, err := InitializeServer(deps)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to load config")
}

func TestInitializeServerMiddlewareInitError(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	mockMiddlewareInit := new(MockMiddlewareInitializer)
	mockProxyService := proxy.NewProxyService()

	config := &config.Config{}
	mockLoader.On("LoadConfig", "./cmd/gateway/config.yaml").Return(config, nil)
	mockMiddlewareInit.On("InitMiddleware", mock.Anything).Return(map[string]server.Middleware{}, errors.New("failed to initialize middleware"))

	deps := ServerDependencies{
		ConfigLoader:          mockLoader,
		MiddlewareInitializer: mockMiddlewareInit,
		Router:                mux.NewRouter(),
		ProxyService:          mockProxyService,
	}

	_, err := InitializeServer(deps)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to initialize middleware")
}
