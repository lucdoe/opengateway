// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/lucdoe/opengateway/internal"
)

const (
	cacheKeyPrefix = "cache:"
	cacheTTL       = 10 * time.Minute

	cacheStatusHeader = "X-Cache"
	cacheHit          = "HIT"
	cacheMiss         = "MISS"
)

// Cache is the part of a store.Store the cache plugin uses. Any error from
// Get counts as a miss.
type Cache interface {
	Get(key string) (string, error)
	Set(key, value string, ttl time.Duration) error
}

type middleware struct {
	store Cache
}

// NewMiddleware returns the cache plugin. It is a shared response cache: one
// stored response is served to every client that sends an equivalent request.
// To keep that safe it only handles GET requests without an Authorization
// header and only stores responses that are not specific to one caller.
func NewMiddleware(store Cache) func(http.Handler) http.Handler {
	return (&middleware{store: store}).wrap
}

func (cm *middleware) wrap(next http.Handler) http.Handler {
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

		answered := upstreamHeader(recorder.Header(), ownHeaders)
		if recorder.StatusCode == http.StatusOK && cacheableResponse(answered) {
			cm.save(key, answered, recorder.Body.Bytes())
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

// upstreamHeader returns the header values added after the earlier plugins
// set theirs: what the upstream answered, without this request's own headers.
func upstreamHeader(header, own http.Header) http.Header {
	added := http.Header{}
	for name, values := range header {
		if name == cacheStatusHeader {
			continue
		}
		if earlier := own[name]; len(earlier) <= len(values) && slices.Equal(values[:len(earlier)], earlier) {
			values = values[len(earlier):]
		}
		if len(values) > 0 {
			added[name] = values
		}
	}
	return added
}

func cacheableResponse(header http.Header) bool {
	if len(header.Values("Set-Cookie")) > 0 {
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

func (cm *middleware) lookup(key string) (entry, bool) {
	value, err := cm.store.Get(key)
	if err != nil {
		return entry{}, false
	}
	return decodeEntry(value)
}

func (cm *middleware) save(key string, header http.Header, body []byte) {
	// A failed write only means the next request is a miss again.
	_ = cm.store.Set(key, encodeEntry(header, body, time.Now()), cacheTTL)
}

func writeCached(w http.ResponseWriter, cached entry) {
	for name, values := range cached.header {
		// After the values earlier plugins set for this request, as on a miss.
		w.Header()[name] = append(w.Header()[name], values...)
	}
	age := int(time.Since(cached.storedAt).Seconds())
	if age < 0 {
		age = 0
	}
	w.Header().Set("Age", strconv.Itoa(age))
	w.Header().Set(cacheStatusHeader, cacheHit)
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, cached.body)
}
