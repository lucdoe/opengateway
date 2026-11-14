// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/plugins/cors"
	"github.com/stretchr/testify/assert"
)

func TestPolicy(t *testing.T) {
	p := cors.New(cors.Config{
		Origins: "http://example.com,http://example.org",
		Methods: "GET, POST",
		Headers: "Content-Type, Authorization",
	})

	assert.True(t, p.AllowsOrigin("http://example.org"), "lists need no space after the comma")
	assert.False(t, p.AllowsOrigin("http://notallowed.com"))
	assert.False(t, p.AllowsOrigin(""))

	assert.True(t, p.AllowsMethod("POST"))
	assert.False(t, p.AllowsMethod("DELETE"))

	assert.True(t, p.AllowsHeaders(""), "no headers requested")
	assert.True(t, p.AllowsHeaders("content-type,authorization"), "browsers send lower case")
	assert.False(t, p.AllowsHeaders("Content-Type, X-Custom"))

	assert.True(t, cors.New(cors.Config{Origins: "http://example.com, *"}).AllowsOrigin("http://any.test"), "* allows every origin")
}

func TestMiddleware(t *testing.T) {
	listed := cors.Config{Origins: "http://allowed.com", Methods: "GET, POST", Headers: "Content-Type"}
	anyOrigin := cors.Config{Origins: "*", Methods: "GET, POST", Headers: "Content-Type"}

	tests := []struct {
		name                                          string
		config                                        cors.Config
		method, origin, requestMethod, requestHeaders string
		status                                        int
		allowOrigin, credentials, vary                string
	}{
		{name: "origin not listed", config: listed, method: "GET", origin: "http://forbidden.com",
			status: http.StatusForbidden, vary: "Origin"},
		{name: "preflight with a method not allowed", config: listed, method: "OPTIONS", origin: "http://allowed.com", requestMethod: "DELETE",
			status: http.StatusForbidden, vary: "Origin"},
		{name: "preflight with a header not allowed", config: listed, method: "OPTIONS", origin: "http://allowed.com", requestMethod: "GET", requestHeaders: "x-custom",
			status: http.StatusForbidden, vary: "Origin"},
		{name: "valid preflight is answered here", config: listed, method: "OPTIONS", origin: "http://allowed.com", requestMethod: "POST", requestHeaders: "content-type",
			status: http.StatusOK, allowOrigin: "http://allowed.com", credentials: "true", vary: "Origin"},
		{name: "listed origin reaches the next handler", config: listed, method: "GET", origin: "http://allowed.com",
			status: http.StatusTeapot, allowOrigin: "http://allowed.com", credentials: "true", vary: "Origin"},
		{name: "no Origin is not a CORS request", config: listed, method: "GET",
			status: http.StatusTeapot, vary: "Origin"},
		{name: "OPTIONS without Origin is not a preflight", config: listed, method: "OPTIONS",
			status: http.StatusTeapot, vary: "Origin"},
		{name: "* allows any origin without credentials", config: anyOrigin, method: "GET", origin: "http://any.test",
			status: http.StatusTeapot, allowOrigin: "*"},
		{name: "* preflight", config: anyOrigin, method: "OPTIONS", origin: "http://any.test", requestMethod: "GET",
			status: http.StatusOK, allowOrigin: "*"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := cors.NewMiddleware(cors.New(tc.config))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}))
			req := httptest.NewRequest(tc.method, "/", nil)
			for name, value := range map[string]string{
				"Origin":                         tc.origin,
				"Access-Control-Request-Method":  tc.requestMethod,
				"Access-Control-Request-Headers": tc.requestHeaders,
			} {
				if value != "" {
					req.Header.Set(name, value)
				}
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
			assert.Equal(t, tc.allowOrigin, rec.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, tc.credentials, rec.Header().Get("Access-Control-Allow-Credentials"))
			assert.Equal(t, tc.vary, rec.Header().Get("Vary"))
			if tc.allowOrigin != "" {
				assert.Equal(t, "GET, POST", rec.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Content-Type", rec.Header().Get("Access-Control-Allow-Headers"))
			} else {
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
			}
		})
	}
}
