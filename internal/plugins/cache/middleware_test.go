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

package cache

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// memoryStore is an in-memory Cache. It records the TTL of the last
// Set and can be made to fail.
type memoryStore struct {
	values  map[string]string
	lastTTL time.Duration
	getErr  error
	setErr  error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{values: map[string]string{}}
}

func (m *memoryStore) Get(key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	value, ok := m.values[key]
	if !ok {
		return "", errors.New("redis: nil")
	}
	return value, nil
}

func (m *memoryStore) Set(key string, value string, expiration time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.values[key] = value
	m.lastTTL = expiration
	return nil
}

// countingUpstream answers every request with respond and counts the calls.
type countingUpstream struct {
	calls   int
	respond func(w http.ResponseWriter, r *http.Request)
}

func (u *countingUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.calls++
	u.respond(w, r)
}

func jsonUpstream() *countingUpstream {
	return &countingUpstream{respond: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}}
}

func get(handler http.Handler, target string, edit ...func(*http.Request)) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", target, nil)
	for _, fn := range edit {
		fn(req)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestCacheServesTheStoredResponseOnAHit(t *testing.T) {
	store := newMemoryStore()
	upstream := &countingUpstream{respond: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Upstream", "yes")
		w.Write([]byte("hello"))
	}}
	handler := NewMiddleware(store)(upstream)

	miss := get(handler, "/api/people")
	hit := get(handler, "/api/people")

	assert.Equal(t, 1, upstream.calls, "the second request must not reach the upstream")
	assert.Equal(t, 10*time.Minute, store.lastTTL)

	assert.Equal(t, "MISS", miss.Header().Get("X-Cache"))
	assert.Equal(t, "HIT", hit.Header().Get("X-Cache"))
	assert.NotEmpty(t, hit.Header().Get("Age"))
	assert.Empty(t, miss.Header().Get("Age"))

	for _, res := range []*httptest.ResponseRecorder{miss, hit} {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "hello", res.Body.String())
		assert.Equal(t, "text/plain; charset=utf-8", res.Header().Get("Content-Type"))
		assert.Equal(t, "yes", res.Header().Get("X-Upstream"))
	}
}

func TestCacheIsSharedBetweenClients(t *testing.T) {
	upstream := jsonUpstream()
	handler := NewMiddleware(newMemoryStore())(upstream)

	get(handler, "/api/people", func(r *http.Request) { r.RemoteAddr = "203.0.113.5:51000" })
	sameClientNewConnection := get(handler, "/api/people", func(r *http.Request) { r.RemoteAddr = "203.0.113.5:51001" })
	otherClient := get(handler, "/api/people", func(r *http.Request) { r.RemoteAddr = "198.51.100.7:40000" })

	assert.Equal(t, 1, upstream.calls)
	assert.Equal(t, "HIT", sameClientNewConnection.Header().Get("X-Cache"))
	assert.Equal(t, "HIT", otherClient.Header().Get("X-Cache"))
}

func TestCacheKeyDependsOnPathQueryAndEncoding(t *testing.T) {
	gzip := func(r *http.Request) { r.Header.Set("Accept-Encoding", "gzip") }

	tests := []struct {
		name      string
		first     string
		second    string
		editFirst func(*http.Request)
		sameEntry bool
	}{
		{name: "same request", first: "/people?page=1", second: "/people?page=1", sameEntry: true},
		{name: "query parameter order", first: "/people?b=2&a=1", second: "/people?a=1&b=2", sameEntry: true},
		{name: "different path", first: "/people", second: "/starships"},
		{name: "different query value", first: "/people?page=1", second: "/people?page=2"},
		{name: "repeated query parameter", first: "/people?tag=a", second: "/people?tag=a&tag=b"},
		{name: "different Accept-Encoding", first: "/people", second: "/people", editFirst: gzip},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			upstream := jsonUpstream()
			handler := NewMiddleware(newMemoryStore())(upstream)

			if tc.editFirst != nil {
				get(handler, tc.first, tc.editFirst)
			} else {
				get(handler, tc.first)
			}
			get(handler, tc.second)

			if tc.sameEntry {
				assert.Equal(t, 1, upstream.calls)
			} else {
				assert.Equal(t, 2, upstream.calls)
			}
		})
	}
}

func TestCacheSkipsRequestsAndResponsesThatMustNotBeShared(t *testing.T) {
	respondWith := func(status int, header, value string) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if header != "" {
				w.Header().Set(header, value)
			}
			w.WriteHeader(status)
			w.Write([]byte("body"))
		}
	}
	ok := respondWith(http.StatusOK, "", "")

	tests := []struct {
		name    string
		method  string
		edit    func(*http.Request)
		respond func(http.ResponseWriter, *http.Request)
	}{
		{name: "POST request", method: "POST", respond: ok},
		{name: "HEAD request", method: "HEAD", respond: ok},
		{name: "request with Authorization", method: "GET", respond: ok,
			edit: func(r *http.Request) { r.Header.Set("Authorization", "Bearer token") }},
		{name: "404 response", method: "GET", respond: respondWith(http.StatusNotFound, "", "")},
		{name: "500 response", method: "GET", respond: respondWith(http.StatusInternalServerError, "", "")},
		{name: "response sets a cookie", method: "GET", respond: respondWith(http.StatusOK, "Set-Cookie", "session=abc")},
		{name: "Cache-Control: no-store", method: "GET", respond: respondWith(http.StatusOK, "Cache-Control", "no-store")},
		{name: "Cache-Control: private", method: "GET", respond: respondWith(http.StatusOK, "Cache-Control", "max-age=60, Private")},
		{name: "Cache-Control: no-cache", method: "GET", respond: respondWith(http.StatusOK, "Cache-Control", "no-cache")},
		{name: "Vary: Cookie", method: "GET", respond: respondWith(http.StatusOK, "Vary", "Accept-Encoding, Cookie")},
		{name: "Vary: *", method: "GET", respond: respondWith(http.StatusOK, "Vary", "*")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newMemoryStore()
			upstream := &countingUpstream{respond: tc.respond}
			handler := NewMiddleware(store)(upstream)

			for i := 0; i < 2; i++ {
				req := httptest.NewRequest(tc.method, "/api/people", nil)
				if tc.edit != nil {
					tc.edit(req)
				}
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				assert.NotEqual(t, "HIT", w.Header().Get("X-Cache"))
			}

			assert.Equal(t, 2, upstream.calls, "both requests must reach the upstream")
			assert.Empty(t, store.values, "nothing may be stored")
		})
	}
}

func TestCacheStoresResponsesThatOnlyVaryOnAcceptEncoding(t *testing.T) {
	upstream := &countingUpstream{respond: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Cache-Control", "public, max-age=60")
		w.Write([]byte("body"))
	}}
	handler := NewMiddleware(newMemoryStore())(upstream)

	get(handler, "/api/people")
	hit := get(handler, "/api/people")

	assert.Equal(t, 1, upstream.calls)
	assert.Equal(t, "HIT", hit.Header().Get("X-Cache"))
}

// Plugins that run before the cache set headers for the current request.
// A cache hit must show the current values, not the ones from the request
// that filled the cache.
func TestCacheKeepsHeadersFromEarlierPluginsFresh(t *testing.T) {
	remaining := []string{"9", "8"}
	origins := []string{"https://a.example", "https://b.example"}
	request := 0

	cached := NewMiddleware(newMemoryStore())(jsonUpstream())
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", remaining[request])
		w.Header().Set("Access-Control-Allow-Origin", origins[request])
		request++
		cached.ServeHTTP(w, r)
	})

	get(handler, "/api/people")
	hit := get(handler, "/api/people")

	assert.Equal(t, "HIT", hit.Header().Get("X-Cache"))
	assert.Equal(t, "8", hit.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "https://b.example", hit.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "application/json", hit.Header().Get("Content-Type"))
}

func TestCacheFallsBackToTheUpstreamWhenTheStoreFails(t *testing.T) {
	t.Run("read fails", func(t *testing.T) {
		store := newMemoryStore()
		upstream := jsonUpstream()
		handler := NewMiddleware(store)(upstream)

		get(handler, "/api/people")
		store.getErr = errors.New("connection refused")
		res := get(handler, "/api/people")

		assert.Equal(t, 2, upstream.calls)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, `{"status":"ok"}`, res.Body.String())
	})

	t.Run("write fails", func(t *testing.T) {
		store := newMemoryStore()
		store.setErr = errors.New("connection refused")
		upstream := jsonUpstream()
		handler := NewMiddleware(store)(upstream)

		res := get(handler, "/api/people")

		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, `{"status":"ok"}`, res.Body.String())
	})

	t.Run("stored entry is unreadable", func(t *testing.T) {
		store := newMemoryStore()
		upstream := jsonUpstream()
		handler := NewMiddleware(store)(upstream)

		get(handler, "/api/people")
		for key := range store.values {
			store.values[key] = "not json"
		}
		res := get(handler, "/api/people")

		assert.Equal(t, 2, upstream.calls)
		assert.Equal(t, "MISS", res.Header().Get("X-Cache"))
	})
}
