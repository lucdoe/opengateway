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

package proxy_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upstream starts a test server that records the last request it received
// and answers with the given status and body.
func upstream(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()
	received := new(http.Request)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := io.ReadAll(r.Body)
		*received = *r
		received.Body = io.NopCloser(strings.NewReader(string(reqBody)))
		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, received
}

func forward(t *testing.T, target string, req *http.Request) *http.Response {
	t.Helper()
	handler, err := proxy.NewHandler(target)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w.Result()
}

func TestHandlerForwardsTheRequestPath(t *testing.T) {
	tests := []struct {
		name         string
		targetPath   string
		requestPath  string
		expectedPath string
	}{
		{"static path", "", "/api/people", "/api/people"},
		{"path parameter value", "", "/api/people/7", "/api/people/7"},
		{"upstream base path is prepended", "/v1", "/api/people/7", "/v1/api/people/7"},
		{"base path with trailing slash", "/v1/", "/api/people", "/v1/api/people"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server, received := upstream(t, http.StatusOK, "ok")

			res := forward(t, server.URL+tc.targetPath, httptest.NewRequest("GET", "http://gateway.test"+tc.requestPath, nil))

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, tc.expectedPath, received.URL.Path)
		})
	}
}

func TestHandlerForwardsQueryMethodAndBody(t *testing.T) {
	server, received := upstream(t, http.StatusOK, "post response")

	req := httptest.NewRequest("POST", "http://gateway.test/api/worlds?page=2&sort=name", strings.NewReader("post body"))
	res := forward(t, server.URL, req)

	assert.Equal(t, "POST", received.Method)
	assert.Equal(t, "2", received.URL.Query().Get("page"))
	assert.Equal(t, "name", received.URL.Query().Get("sort"))
	body, _ := io.ReadAll(received.Body)
	assert.Equal(t, "post body", string(body))

	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, "post response", string(responseBody))
}

func TestHandlerAddressesTheUpstreamHost(t *testing.T) {
	server, received := upstream(t, http.StatusOK, "ok")
	upstreamURL, _ := url.Parse(server.URL)

	req := httptest.NewRequest("GET", "http://gateway.test/api/people", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	forward(t, server.URL, req)

	assert.Equal(t, upstreamURL.Host, received.Host)
	assert.Equal(t, "gateway.test", received.Header.Get("X-Forwarded-Host"))
	assert.Equal(t, "192.0.2.1", received.Header.Get("X-Forwarded-For"), "client-supplied value must be replaced by the real peer address")
}

func TestHandlerReturnsTheUpstreamResponse(t *testing.T) {
	server, _ := upstream(t, http.StatusInternalServerError, "server error")

	res := forward(t, server.URL, httptest.NewRequest("GET", "http://gateway.test/api/people", nil))

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "yes", res.Header.Get("X-Upstream"))
	body, _ := io.ReadAll(res.Body)
	assert.Equal(t, "server error", string(body))
}

func TestHandlerAnswersBadGatewayWhenUpstreamIsDown(t *testing.T) {
	server, _ := upstream(t, http.StatusOK, "ok")
	target := server.URL
	server.Close()

	res := forward(t, target, httptest.NewRequest("GET", "http://gateway.test/api/people", nil))

	assert.Equal(t, http.StatusBadGateway, res.StatusCode)
}

func TestNewHandlerRejectsInvalidTargets(t *testing.T) {
	for _, target := range []string{"", ":", "swapi.dev", "/api", "ftp://swapi.dev", "http://"} {
		t.Run(target, func(t *testing.T) {
			handler, err := proxy.NewHandler(target)

			assert.Nil(t, handler)
			assert.ErrorContains(t, err, "invalid upstream URL")
		})
	}
}
