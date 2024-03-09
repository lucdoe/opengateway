package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/middleware"
)

const (
	TestMethods = "GET,POST"
	TestHeaders = "Content-Type"
)

func TestCORSHandlerResponseHeaders(t *testing.T) {
	c := middleware.CORS{
		AccessControlAllowOrigin:   "*",
		AccessControlAllowMethods:  TestMethods,
		AccessControlAllowHeaders:  TestHeaders,
		AccessControlExposeHeaders: "Authorization",
		AccessControlMaxAge:        "3600",
	}
	handler := middleware.CORSHandler(c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":   "*",
		"Access-Control-Allow-Methods":  TestMethods,
		"Access-Control-Allow-Headers":  TestHeaders,
		"Access-Control-Expose-Headers": "Authorization",
		"Access-Control-Max-Age":        "3600",
	}

	for key, expectedValue := range expectedHeaders {
		if value := w.Header().Get(key); value != expectedValue {
			t.Errorf("Header %s: expected %s, got %s", key, expectedValue, value)
		}
	}
}

func TestCORSHandlerForOptionsMethod(t *testing.T) {
	c := middleware.CORS{
		AccessControlAllowOrigin:   "*",
		AccessControlAllowMethods:  TestMethods,
		AccessControlAllowHeaders:  TestHeaders,
		AccessControlExposeHeaders: "Authorization",
		AccessControlMaxAge:        "3600",
	}

	handler := middleware.CORSHandler(c)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for OPTIONS method")
	}))

	req := httptest.NewRequest("OPTIONS", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}
