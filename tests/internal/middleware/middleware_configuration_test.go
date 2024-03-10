package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/middleware"
)

type mockHandler struct {
	called bool
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
}

func TestMiddlewareHandlerAdd(t *testing.T) {
	handler := middleware.MiddlewareHandlerConstructor()

	middlewares := []config.Middleware{
		{
			Name: "Logger",
		},
		{
			Name: "Unknown",
		},
	}

	mock := &mockHandler{}

	decoratedHandler := handler.Add(mock, middlewares)

	responseWriter := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)

	decoratedHandler.ServeHTTP(responseWriter, request)

	if !mock.called {
		t.Errorf("Expected the mock handler to be called")
	}
}
