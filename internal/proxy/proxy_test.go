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
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/stretchr/testify/assert"
)

const (
	exampleURL = "http://example.com"
)

func TestProxyReverseProxy(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))
	defer testServer.Close()

	proxyService := proxy.NewProxyService()

	req := httptest.NewRequest("GET", exampleURL, nil)
	w := httptest.NewRecorder()

	err := proxyService.ReverseProxy(testServer.URL, w, req)
	assert.NoError(t, err)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, "hello world", string(responseBody))
}

func TestProxyReverseProxyWithPost(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "post body", string(body))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("post response"))
	}))
	defer testServer.Close()

	proxyService := proxy.NewProxyService()

	req := httptest.NewRequest("POST", exampleURL, io.NopCloser(io.MultiReader(bytes.NewReader([]byte("post body")))))
	w := httptest.NewRecorder()

	err := proxyService.ReverseProxy(testServer.URL, w, req)
	assert.NoError(t, err)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, "post response", string(responseBody))
}

func TestProxyReverseProxyWithQueryParams(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "value", r.URL.Query().Get("param"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("query response"))
	}))
	defer testServer.Close()

	proxyService := proxy.NewProxyService()

	req := httptest.NewRequest("GET", "http://example.com?param=value", nil)
	w := httptest.NewRecorder()

	err := proxyService.ReverseProxy(testServer.URL, w, req)
	assert.NoError(t, err)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, "query response", string(responseBody))
}

func TestProxyReverseProxyHandlesInvalidURL(t *testing.T) {
	proxyService := proxy.NewProxyService()

	req := httptest.NewRequest("GET", exampleURL, nil)
	w := httptest.NewRecorder()

	err := proxyService.ReverseProxy(":", w, req)
	assert.Error(t, err)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestProxyReverseProxyHandlesServerError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer testServer.Close()

	proxyService := proxy.NewProxyService()

	req := httptest.NewRequest("GET", exampleURL, nil)
	w := httptest.NewRecorder()

	err := proxyService.ReverseProxy(testServer.URL, w, req)
	assert.NoError(t, err)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, "server error", string(responseBody))
}
