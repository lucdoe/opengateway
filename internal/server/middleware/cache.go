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

package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lucdoe/opengateway/internal"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
)

const (
	cacheKeyPrefix = "cache:"
	cacheTTL       = 10 * time.Minute

	cacheStatusHeader = "X-Cache"
	cacheHit          = "HIT"
	cacheMiss         = "MISS"
)

// CacheMiddleware is a shared response cache: one stored response is served
// to every client that sends an equivalent request. To keep that safe it only
// handles GET requests without an Authorization header and only stores
// responses that are not specific to one caller.
type CacheMiddleware struct {
	Cache cache.Cache
}

func NewCacheMiddleware(c cache.Cache) *CacheMiddleware {
	return &CacheMiddleware{Cache: c}
}

// cachedResponse is what is kept in the store for one cache key.
type cachedResponse struct {
	Header   http.Header `json:"header"`
	Body     []byte      `json:"body"`
	StoredAt time.Time   `json:"stored_at"`
}

func (cm *CacheMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cacheableRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		key := cacheKey(r)
		if cached, ok := cm.lookup(key); ok {
			writeCached(w, cached)
			return
		}

		// Headers already on the response were put there by plugins that run
		// before the cache, such as cors and rate-limit. They belong to this
		// request, so they are not stored with the response.
		ownHeaders := w.Header().Clone()

		w.Header().Set(cacheStatusHeader, cacheMiss)
		recorder := internal.NewResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if cacheableResponse(recorder) {
			cm.store(key, recorder, ownHeaders)
		}
	})
}

// cacheKey identifies equivalent requests. Query parameters are sorted, so
// their order does not matter. Accept-Encoding is part of the key because the
// upstream may answer with a different encoding depending on it.
func cacheKey(r *http.Request) string {
	return cacheKeyPrefix + r.Method + ":" + r.URL.Path + "?" + r.URL.Query().Encode() +
		":" + strings.TrimSpace(r.Header.Get("Accept-Encoding"))
}

func cacheableRequest(r *http.Request) bool {
	return r.Method == http.MethodGet && r.Header.Get("Authorization") == ""
}

func cacheableResponse(recorder *internal.ResponseRecorder) bool {
	header := recorder.Header()
	if recorder.StatusCode != http.StatusOK || len(header.Values("Set-Cookie")) > 0 {
		return false
	}
	for _, directive := range headerTokens(header, "Cache-Control") {
		name, _, _ := strings.Cut(directive, "=")
		if name == "no-store" || name == "no-cache" || name == "private" {
			return false
		}
	}
	for _, field := range headerTokens(header, "Vary") {
		if field != "accept-encoding" {
			return false
		}
	}
	return true
}

// headerTokens returns the lower-cased, comma-separated values of a header.
func headerTokens(header http.Header, name string) []string {
	var tokens []string
	for _, value := range header.Values(name) {
		for _, token := range strings.Split(value, ",") {
			if token = strings.ToLower(strings.TrimSpace(token)); token != "" {
				tokens = append(tokens, token)
			}
		}
	}
	return tokens
}

func (cm *CacheMiddleware) lookup(key string) (cachedResponse, bool) {
	var cached cachedResponse
	value, err := cm.Cache.Get(key)
	if err != nil || value == "" {
		return cached, false
	}
	if err := json.Unmarshal([]byte(value), &cached); err != nil {
		return cached, false
	}
	return cached, true
}

func (cm *CacheMiddleware) store(key string, recorder *internal.ResponseRecorder, exclude http.Header) {
	header := http.Header{}
	for name, values := range recorder.Header() {
		if _, excluded := exclude[name]; excluded || name == cacheStatusHeader {
			continue
		}
		header[name] = values
	}

	value, err := json.Marshal(cachedResponse{
		Header:   header,
		Body:     recorder.Body.Bytes(),
		StoredAt: time.Now(),
	})
	if err != nil {
		return
	}
	// A failed write only means the next request is a miss again.
	_ = cm.Cache.Set(key, string(value), cacheTTL)
}

func writeCached(w http.ResponseWriter, cached cachedResponse) {
	for name, values := range cached.Header {
		// Headers set by earlier plugins for this request win over stored ones.
		if _, set := w.Header()[name]; !set {
			w.Header()[name] = values
		}
	}
	age := int(time.Since(cached.StoredAt).Seconds())
	if age < 0 {
		age = 0
	}
	w.Header().Set("Age", strconv.Itoa(age))
	w.Header().Set(cacheStatusHeader, cacheHit)
	w.WriteHeader(http.StatusOK)
	w.Write(cached.Body)
}
