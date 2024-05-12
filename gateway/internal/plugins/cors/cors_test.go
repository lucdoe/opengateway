package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cors"
)

func TestCorsMiddlewareOriginValidation(t *testing.T) {
	tests := []struct {
		name         string
		origin       string
		expectedPass bool
	}{
		{"Allowed Origin", "http://example.com", true},
		{"Disallowed Origin", "http://notallowed.com", false},
		{"Wildcard Origin", "*", true},
	}

	corsConfig := cors.CORSConfig{
		Origins: "http://example.com, http://example.org, *",
		Methods: "GET, POST",
		Headers: "Content-Type, Authorization",
	}
	corsMiddleware := cors.NewCors(corsConfig)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://localhost/", nil)
			req.Header.Set("Origin", tc.origin)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			corsHandler := corsMiddleware.Middleware(handler)
			corsHandler.ServeHTTP(rr, req)

			if (rr.Code == http.StatusOK) != tc.expectedPass {
				t.Errorf("Expected pass: %v, got: %v, for origin: %v", tc.expectedPass, rr.Code == http.StatusOK, tc.origin)
			}
		})
	}
}

func TestCorsMiddlewareMethodValidation(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Origins: "*",
		Methods: "GET, POST",
		Headers: "Content-Type",
	}
	corsMiddleware := cors.NewCors(corsConfig)

	req := httptest.NewRequest("OPTIONS", "http://localhost/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := corsMiddleware.Middleware(handler)
	corsHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected StatusForbidden for method DELETE, got: %v", rr.Code)
	}
}

func TestCorsMiddlewareHeaderValidation(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Origins: "*",
		Methods: "GET, POST",
		Headers: "Content-Type",
	}
	corsMiddleware := cors.NewCors(corsConfig)

	req := httptest.NewRequest("OPTIONS", "http://localhost/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := corsMiddleware.Middleware(handler)
	corsHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected StatusForbidden for header X-Custom-Header, got: %v", rr.Code)
	}
}

func TestCorsMiddlewareCredentials(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Origins: "*",
		Methods: "GET, POST",
		Headers: "Content-Type",
	}
	corsMiddleware := cors.NewCors(corsConfig)

	req := httptest.NewRequest("GET", "http://localhost/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := corsMiddleware.Middleware(handler)
	corsHandler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected true for Access-Control-Allow-Credentials, got: %v", rr.Header().Get("Access-Control-Allow-Credentials"))
	}
}
