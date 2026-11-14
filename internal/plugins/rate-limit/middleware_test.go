// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeCounter counts every key together and records the keys it was given.
type fakeCounter struct {
	count   int64
	resetIn time.Duration
	err     error
	keys    []string
}

func (f *fakeCounter) Increment(key string, _ time.Duration) (int64, time.Duration, error) {
	f.keys = append(f.keys, key)
	if f.err != nil {
		return 0, 0, f.err
	}
	f.count++
	return f.count, f.resetIn, nil
}

func serve(handler http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/", nil)
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

var ok = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func TestMiddlewareCountsDownThenRejects(t *testing.T) {
	handler := NewMiddleware(NewLimiter(&fakeCounter{resetIn: 1500 * time.Millisecond}, 2, time.Minute), ByClientIP)(ok)

	for _, remaining := range []string{"1", "0"} {
		rec := serve(handler, "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, remaining, rec.Header().Get("X-RateLimit-Remaining"))
		assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Reset"), "rounded up to whole seconds")
		assert.Empty(t, rec.Header().Get("Retry-After"))
	}

	rec := serve(handler, "")
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "2", rec.Header().Get("Retry-After"))
	assert.JSONEq(t, `{"message": "Rate limit exceeded"}`, rec.Body.String())
}

func TestMiddlewareCountsPerClientIP(t *testing.T) {
	counter := &fakeCounter{resetIn: time.Minute}
	handler := NewMiddleware(NewLimiter(counter, 10, time.Minute), ByClientIP)(ok)

	for _, addr := range []string{"203.0.113.5:51000", "203.0.113.5:51001", "[2001:db8::1]:51000", "[2001:db8::2]:51000", "[2001:db8:0:1::1]:51000"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = addr
		req.Header.Set("X-Forwarded-For", "10.0.0.1") // ignored: clients can set it
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	assert.Equal(t, []string{
		"ratelimit:203.0.113.5", "ratelimit:203.0.113.5",
		"ratelimit:2001:db8::/64", "ratelimit:2001:db8::/64", // one /64, one client
		"ratelimit:2001:db8:0:1::/64",
	}, counter.keys)
}

// A store outage must not take down the services behind the plugin.
func TestMiddlewareLetsRequestsThroughWhenTheStoreFails(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	reached := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached++ })
	// Same text as the limit error, but not that error.
	handler := NewMiddleware(NewLimiter(&fakeCounter{err: errors.New("rate limit exceeded")}, 10, time.Minute), ByClientIP)(next)

	var last *httptest.ResponseRecorder
	for i := 0; i < 3; i++ {
		last = serve(handler, "")
	}

	assert.Equal(t, 3, reached, "every request is forwarded")
	assert.Equal(t, http.StatusOK, last.Code)
	assert.Empty(t, last.Header().Get("X-RateLimit-Limit"), "no limit is claimed that was not checked")
	assert.Equal(t, 1, strings.Count(logs.String(), "rate-limit: store failed"), "the outage is logged once, not per request")
	assert.Contains(t, logs.String(), "rate limit exceeded")
}

func TestClientKey(t *testing.T) {
	for ip, want := range map[string]string{
		"203.0.113.5":          "203.0.113.5",
		"::ffff:203.0.113.5":   "::ffff:203.0.113.5", // IPv4 client on an IPv6 socket
		"2001:db8:1:2:3:4:5:6": "2001:db8:1:2::/64",
		"not an address":       "not an address",
		"":                     "",
	} {
		assert.Equal(t, want, clientKey(ip), ip)
	}
}

func TestByHeaderCountsPerValueAndFallsBackToTheIP(t *testing.T) {
	key := ByHeader("CF-Connecting-IP")
	request := func(remoteAddr, header string) *http.Request {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = remoteAddr
		if header != "" {
			req.Header.Set("CF-Connecting-IP", header)
		}
		return req
	}

	alice := key(request("10.0.0.1:1", "198.51.100.7"))
	assert.Regexp(t, `^header:[0-9a-f]{32}$`, alice, "the value is stored hashed")
	assert.NotContains(t, alice, "198.51.100.7")
	assert.Equal(t, alice, key(request("10.0.0.2:1", "198.51.100.7")), "the same value from another proxy address is the same client")
	assert.NotEqual(t, alice, key(request("10.0.0.1:1", "198.51.100.8")))
	assert.Equal(t, "10.0.0.1", key(request("10.0.0.1:1", "")), "without the header, the client IP counts")
}

func TestByValueKeepsKindsApart(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	value := func(*http.Request) string { return "same" }

	assert.NotEqual(t, ByValue("sub", value)(req), ByValue("header", value)(req))
}

func TestMiddlewareCountsByTheGivenKey(t *testing.T) {
	counter := &fakeCounter{}
	handler := NewMiddleware(NewLimiter(counter, 10, time.Minute), func(*http.Request) string { return "tenant-7" })(ok)

	serve(handler, "203.0.113.5:1")

	assert.Equal(t, []string{"ratelimit:tenant-7"}, counter.keys)
}
