// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	handler, err := proxy.NewHandler(target, time.Second)
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
	assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
}

// slowUpstream starts a test server that only answers once release is closed.
func slowUpstream(t *testing.T) (target string, release chan struct{}) {
	t.Helper()
	release = make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
		w.Write([]byte("late"))
	}))
	t.Cleanup(server.Close)
	return server.URL, release
}

func TestHandlerAnswersGatewayTimeoutWhenUpstreamIsTooSlow(t *testing.T) {
	target, release := slowUpstream(t)
	defer close(release)

	handler, err := proxy.NewHandler(target, 50*time.Millisecond)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	started := time.Now()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "http://gateway.test/api/people", nil))

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.JSONEq(t, `{"message": "Gateway timeout"}`, w.Body.String())
	assert.Less(t, time.Since(started), 2*time.Second)
}

func TestHandlerWaitsWithinTheTimeout(t *testing.T) {
	target, release := slowUpstream(t)
	time.AfterFunc(50*time.Millisecond, func() { close(release) })

	handler, err := proxy.NewHandler(target, 5*time.Second)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "http://gateway.test/api/people", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "late", w.Body.String())
}

func TestNewHandlerRejectsInvalidTargets(t *testing.T) {
	for _, target := range []string{"", ":", "swapi.dev", "/api", "ftp://swapi.dev", "http://"} {
		t.Run(target, func(t *testing.T) {
			handler, err := proxy.NewHandler(target, time.Second)

			assert.Nil(t, handler)
			assert.ErrorContains(t, err, "invalid upstream URL")
		})
	}
}

// With the default of 2 idle connections per host, every wave of concurrent
// requests would dial most of its connections again.
func TestHandlerReusesUpstreamConnections(t *testing.T) {
	var dials atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond) // keep the wave's requests in flight together
		w.Write([]byte("ok"))
	}))
	server.Config.ConnState = func(c net.Conn, state http.ConnState) {
		if state == http.StateNew {
			dials.Add(1)
		}
	}
	server.Start()
	t.Cleanup(server.Close)

	handler, err := proxy.NewHandler(server.URL, time.Second)
	require.NoError(t, err)

	const concurrency, waves = 20, 3
	for wave := 0; wave < waves; wave++ {
		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", "http://gateway.test/api/people", nil))
				assert.Equal(t, http.StatusOK, w.Code)
			}()
		}
		wg.Wait()
	}

	assert.LessOrEqual(t, int(dials.Load()), concurrency+2, "later waves reuse the first wave's connections")
}

// Bodies larger than one copy buffer arrive intact while many responses share
// the pooled buffers at once.
func TestHandlerCopiesLargeBodiesThroughPooledBuffers(t *testing.T) {
	bodies := make([][]byte, 8)
	for i := range bodies {
		bodies[i] = bytes.Repeat([]byte{byte('a' + i)}, 200*1024+i)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i, _ := strconv.Atoi(r.URL.Query().Get("i"))
		w.Write(bodies[i])
	}))
	t.Cleanup(server.Close)

	handler, err := proxy.NewHandler(server.URL, time.Second)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for round := 0; round < 5; round++ {
		for i := range bodies {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", "http://gateway.test/big?i="+strconv.Itoa(i), nil))
				assert.True(t, bytes.Equal(bodies[i], w.Body.Bytes()), "body %d arrived intact", i)
			}(i)
		}
	}
	wg.Wait()
}

func TestHandlerKeepsOneRequestIDWhenTheUpstreamEchoesIt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
	}))
	t.Cleanup(server.Close)
	handler, err := proxy.NewHandler(server.URL, time.Second)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "http://gateway.test/", nil)
	req.Header.Set("X-Request-ID", "abc")
	w := httptest.NewRecorder()
	w.Header().Set("X-Request-ID", "abc") // as the request ID middleware does
	handler.ServeHTTP(w, req)

	assert.Equal(t, []string{"abc"}, w.Header().Values("X-Request-ID"))
}
