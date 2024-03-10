package handler_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/handler"
)

type mockURLConstructor struct{}

func (m *mockURLConstructor) ConstructURL(baseURL, path string) (*url.URL, error) {
	return url.Parse("http://example.com")
}

type mockMiddlewareHandler struct{}

func (m *mockMiddlewareHandler) Add(handler http.Handler, middlewares []config.Middleware) http.Handler {
	return handler
}

type mockServer struct {
	ListenAndServeCalled bool
	Addr                 string
	Handler              http.Handler
}

func (m *mockServer) ListenAndServe(addr string, handler http.Handler) error {
	m.ListenAndServeCalled = true
	m.Addr = addr
	m.Handler = handler
	return nil
}

func TestGatewayHandler(t *testing.T) {
	mux := http.NewServeMux()
	urlConstructor := &mockURLConstructor{}
	middlewareHandler := &mockMiddlewareHandler{}
	server := &mockServer{}

	cfg := &config.TopLevel{}

	handler.GatewayHandler(cfg, mux, urlConstructor, middlewareHandler, server)

	if !server.ListenAndServeCalled {
		t.Errorf("expected ListenAndServe to be called")
	}
}
