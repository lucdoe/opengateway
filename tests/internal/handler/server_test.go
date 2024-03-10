package handler_test

import (
	"net/http"
	"testing"

	"github.com/lucdoe/opengateway/internal/handler"
)

type MockServer struct {
	ListenAndServeCalled bool
	Addr                 string
	Handler              http.Handler
}

func (m *MockServer) ListenAndServe(addr string, handler http.Handler) error {
	m.ListenAndServeCalled = true
	m.Addr = addr
	m.Handler = handler
	return nil
}

func TestServerStart(t *testing.T) {
	startServer := func(server handler.ServerI) error {
		addr := ":8080"
		dummyHandler := http.NewServeMux()
		return server.ListenAndServe(addr, dummyHandler)
	}

	mockServer := &MockServer{}

	err := startServer(mockServer)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !mockServer.ListenAndServeCalled {
		t.Errorf("ListenAndServe was not called on the mock server")
	}
}
