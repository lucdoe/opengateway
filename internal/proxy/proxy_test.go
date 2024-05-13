package proxy_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/open-gateway/gateway/internal/proxy"
	"github.com/stretchr/testify/assert"
)

func TestProxyReverseProxy(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))
	defer testServer.Close()

	proxyService := proxy.NewProxyService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	err := proxyService.ReverseProxy(testServer.URL, w, req)
	assert.NoError(t, err)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, "hello world", string(responseBody))
}
