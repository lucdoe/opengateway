package router_test

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/lucdoe/opengateway/internal/router"
)

type MockMux struct {
	RegisteredHandlers map[string]http.Handler
}

func NewMockMux() *MockMux {
	return &MockMux{RegisteredHandlers: make(map[string]http.Handler)}
}

func (m *MockMux) Handle(pattern string, handler http.Handler) {
	m.RegisteredHandlers[pattern] = handler
}

type MockURLConstructor struct{}

func (m *MockURLConstructor) ConstructURL(baseURL, path string) (*url.URL, error) {
	if baseURL == "http://fail.example.com" {
		return nil, errors.New("failed to construct URL")
	}
	return url.Parse(baseURL + path)
}

type MockMiddlewareHandler struct{}

func (m *MockMiddlewareHandler) Add(handler http.Handler, middlewares []config.Middleware) http.Handler {
	return handler
}

func TestRegisterServices(t *testing.T) {
	mockMux := NewMockMux()

	mockURLConstructor := &MockURLConstructor{}
	mockMiddlewareHandler := &MockMiddlewareHandler{}
	serviceRouter := router.ServiceRouterConstructor(mockMux, mockURLConstructor, mockMiddlewareHandler)

	services := []config.Service{
		{
			Protocol: "http",
			Host:     "example.com",
			Port:     80,
			BasePath: "/service",
			Endpoints: []config.Endpoint{
				{Path: "/endpoint1", Middlewares: []config.Middleware{}},
			},
		},
		{
			Protocol: "http",
			Host:     "fail.example.com",
			Port:     80,
			BasePath: "/fail-service",
			Endpoints: []config.Endpoint{
				{Path: "/fail-endpoint", Middlewares: []config.Middleware{}},
			},
		},
	}

	serviceRouter.RegisterServices(services)

	expectedPath := "/service/endpoint1"
	if _, ok := mockMux.RegisteredHandlers[expectedPath]; !ok {
		t.Errorf("Expected path %s to be registered, but it was not", expectedPath)
	}

	expectedHandlerType := reflect.TypeOf(proxy.ReverseProxyHandler(nil)).String()
	registeredHandlerType := reflect.TypeOf(mockMux.RegisteredHandlers[expectedPath]).String()

	if registeredHandlerType != expectedHandlerType {
		t.Errorf("Expected handler of type %s, got %s", expectedHandlerType, registeredHandlerType)
	}
}
