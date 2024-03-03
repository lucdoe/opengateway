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

func TestCORSConfigurations(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(c *middleware.CORS)
		checkOrigin     string
		expectedError   bool
		expectedOrigin  string
		expectedMethods string
	}{
		{
			name: "SetAccessControlAllowOrigin with specific origin",
			setupFunc: func(c *middleware.CORS) {
				c.SetAccessControlAllowOrigin("http://example.com")
			},
			expectedOrigin: "http://example.com",
		},
		{
			name: "SetAccessControlAllowOrigin with empty origin",
			setupFunc: func(c *middleware.CORS) {
				c.SetAccessControlAllowOrigin("")
			},
			expectedOrigin: "*",
		},
		{
			name: "SetAccessControlAllowMethods with valid methods",
			setupFunc: func(c *middleware.CORS) {
				err := c.SetAccessControlAllowMethods([]string{"GET", "POST"})
				if err != nil {
					t.Fatal("Setup error: ", err)
				}
			},
			expectedMethods: "GET,POST",
		},
		{
			name: "SetAccessControlAllowMethods with invalid method",
			setupFunc: func(c *middleware.CORS) {
				err := c.SetAccessControlAllowMethods([]string{"INVALID"})
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
			},
			expectedError: true,
		},
		{
			name: "SetAccessControlMaxAge with valid value",
			setupFunc: func(c *middleware.CORS) {
				c.SetAccessControllMaxAge("3600")
			},
			expectedOrigin: "3600",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &middleware.CORS{}

			if tc.setupFunc != nil {
				tc.setupFunc(c)
			}

			if c.AccessControlAllowOrigin != tc.expectedOrigin {
				t.Errorf("Expected AccessControlAllowOrigin %s, got %s", tc.expectedOrigin, c.AccessControlAllowOrigin)
			}

			if !tc.expectedError && c.AccessControlAllowMethods != tc.expectedMethods {
				t.Errorf("Expected AccessControlAllowMethods %s, got %s", tc.expectedMethods, c.AccessControlAllowMethods)
			}
		})
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
