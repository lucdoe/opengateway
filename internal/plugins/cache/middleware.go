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
)

const (
	cacheKeyPrefix = "cache:"
	// cacheTTL is the longest a response is kept, and how long one without
	// max-age or s-maxage is kept.
	cacheTTL = 10 * time.Minute

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
// To keep that safe it only handles GET requests without a Cookie header and
// only stores responses that are not specific to one caller, no longer than
// their Cache-Control allows. A request with an Authorization header is only
// answered from, and stored in, the cache when the response says it may be
// shared anyway, with public or s-maxage (RFC 9111, section 3.5).
func NewMiddleware(store Cache) func(http.Handler) http.Handler {
	return (&middleware{store: store}).wrap
}

func (cm *middleware) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cacheableRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		// The auth plugin runs before the cache, so the caller has been
		// checked. What is left to rule out is a response meant for one user.
		authorized := r.Header.Get("Authorization") != ""
		shareable := func(header http.Header) bool { return !authorized || sharedDespiteAuthorization(header) }

		key := cacheKey(r)
		if cached, ok := cm.lookup(key); ok && shareable(cached.header) {
			writeCached(w, cached)
			return
		}

		// Headers already on the response were put there by plugins that run
		// before the cache, such as cors and rate-limit. They belong to this
		// request, so they are not stored with the response.
		ownHeaders := w.Header().Clone()

		w.Header().Set(cacheStatusHeader, cacheMiss)
		rec := &recorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		if rec.status != http.StatusOK || rec.tooBig || rec.streamed {
			return
		}
		answered := upstreamHeader(rec.Header(), ownHeaders)
		if ttl, ok := storeFor(answered); ok && shareable(answered) {
			cm.save(key, answered, rec.body.Bytes(), ttl)
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
	// A cookie can make the upstream answer for one user only, without saying
	// so in Vary, so such requests are never shared.
	return r.Method == http.MethodGet && r.Header.Get("Cookie") == ""
}

// sharedDespiteAuthorization reports whether a response to a request with
// an Authorization header may be shared with other callers: only when its
// Cache-Control says public or s-maxage. must-revalidate would allow it too,
// but the cache does not revalidate.
func sharedDespiteAuthorization(header http.Header) bool {
	for _, directive := range headerTokens(header, "Cache-Control") {
		if name, _, _ := strings.Cut(directive, "="); name == "public" || name == "s-maxage" {
			return true
		}
	}
	return false
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

// storeFor reports whether a response with header may be shared, and for
// how long: s-maxage, else max-age, at most cacheTTL.
func storeFor(header http.Header) (time.Duration, bool) {
	if len(header.Values("Set-Cookie")) > 0 {
		return 0, false
	}
	for _, field := range headerTokens(header, "Vary") {
		if field != "accept-encoding" {
			return 0, false
		}
	}
	maxAge, sMaxAge := -1, -1
	for _, directive := range headerTokens(header, "Cache-Control") {
		name, value, _ := strings.Cut(directive, "=")
		switch name {
		case "no-store", "no-cache", "private":
			return 0, false
		case "max-age":
			maxAge = seconds(value)
		case "s-maxage":
			sMaxAge = seconds(value)
		}
	}
	age := sMaxAge
	if age < 0 {
		age = maxAge
	}
	switch {
	case age == 0:
		return 0, false
	case age < 0 || age > int(cacheTTL/time.Second):
		return cacheTTL, true
	}
	return time.Duration(age) * time.Second, true
}

// seconds parses a delta-seconds value, or returns -1 if it is not one.
func seconds(value string) int {
	n, err := strconv.Atoi(strings.Trim(value, `"`))
	if err != nil || n < 0 {
		return -1
	}
	return n
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

func (cm *middleware) save(key string, header http.Header, body []byte, ttl time.Duration) {
	// A failed write only means the next request is a miss again.
	_ = cm.store.Set(key, encodeEntry(header, body, time.Now()), ttl)
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
