// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	handler := cors.NewMiddleware(cors.New(cors.Config{
		Origins: "http://allowed.com",
		Methods: "GET, POST",
		Headers: "Content-Type",
	}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) }))

	tests := []struct {
		name, method, origin, requestMethod, requestHeaders string
		status                                              int
	}{
		{"invalid origin", "GET", "http://forbidden.com", "", "", http.StatusForbidden},
		{"preflight with a method not allowed", "OPTIONS", "http://allowed.com", "DELETE", "", http.StatusForbidden},
		{"preflight with a header not allowed", "OPTIONS", "http://allowed.com", "GET", "x-custom", http.StatusForbidden},
		{"valid preflight is answered here", "OPTIONS", "http://allowed.com", "POST", "content-type", http.StatusOK},
		{"valid request reaches the next handler", "GET", "http://allowed.com", "", "", http.StatusTeapot},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", tc.requestMethod)
			req.Header.Set("Access-Control-Request-Headers", tc.requestHeaders)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
			if tc.status != http.StatusForbidden {
				assert.Equal(t, tc.origin, rec.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "GET, POST", rec.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Content-Type", rec.Header().Get("Access-Control-Allow-Headers"))
			}
		})
	}
}
