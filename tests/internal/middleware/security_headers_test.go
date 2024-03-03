package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/middleware"
)

func TestSecurityHeadersHandler(t *testing.T) {
	securityTestHeaders := middleware.SecurityHeaders{
		ContentSecurityPolicy: "default-src 'self'; frame-ancestors 'none'; script-src 'self'; object-src 'none'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;",
		ContentTypeOptions:    "nosniff",
		FrameOptions:          "DENY",
		XSSProtection:         "1; mode=block",
		StrictTransport:       "max-age=63072000; includeSubDomains",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		DNSPrefetchControl:    "off",
		FeaturePolicy:         "geolocation 'none'; midi 'none'; notifications 'none'; push 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; speaker 'none'; vibrate 'none'; fullscreen 'self'; payment 'none';",
		PermissionsPolicy:     "geolocation=(), midi=(), notifications=(), push=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), payment=()",
	}

	handler := middleware.SecurityHeadersHandler(securityTestHeaders)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//
	}))

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	expectedTestHeaders := map[string]string{
		"Content-Security-Policy":   "default-src 'self'; frame-ancestors 'none'; script-src 'self'; object-src 'none'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"X-DNS-Prefetch-Control":    "off",
		"Feature-Policy":            "geolocation 'none'; midi 'none'; notifications 'none'; push 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; speaker 'none'; vibrate 'none'; fullscreen 'self'; payment 'none';",
		"Permissions-Policy":        "geolocation=(), midi=(), notifications=(), push=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), payment=()",
	}

	for key, expectedValue := range expectedTestHeaders {
		if value := w.Header().Get(key); value != expectedValue {
			t.Errorf("Header %s: expected %s, got %s", key, expectedValue, value)
		}
	}
}

func TestDefaultSecurityHeaders(t *testing.T) {
	expectedDefaultHeaders := middleware.SecurityHeaders{
		ContentSecurityPolicy: "default-src 'self'; frame-ancestors 'none'; script-src 'self'; object-src 'none'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;",
		ContentTypeOptions:    "nosniff",
		FrameOptions:          "DENY",
		XSSProtection:         "1; mode=block",
		StrictTransport:       "max-age=63072000; includeSubDomains",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		DNSPrefetchControl:    "off",
		FeaturePolicy:         "geolocation 'none'; midi 'none'; notifications 'none'; push 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; speaker 'none'; vibrate 'none'; fullscreen 'self'; payment 'none';",
		PermissionsPolicy:     "geolocation=(), midi=(), notifications=(), push=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), payment=()",
	}

	headers := middleware.SecurityHeaders{}
	if got := headers.DefaultSecurityHeaders(); got != expectedDefaultHeaders {
		t.Errorf("DefaultSecurityHeaders: expected %v, got %v", expectedDefaultHeaders, got)
	}
}
