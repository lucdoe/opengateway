// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package requestid

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// serve runs a request with the incoming header value (none when empty)
// and returns the ID the handler saw and the response.
func serve(incoming string) (seen string, rec *httptest.ResponseRecorder) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get(Header)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	if incoming != "" {
		req.Header.Set(Header, incoming)
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return seen, rec
}

var generated = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestMiddlewareSetsANewIDWhenNoneIsSent(t *testing.T) {
	seen, rec := serve("")

	assert.Regexp(t, generated, seen)
	assert.Equal(t, seen, rec.Header().Get(Header), "the client gets the ID the upstream got")

	other, _ := serve("")
	assert.NotEqual(t, seen, other, "every request gets its own ID")
}

func TestMiddlewareKeepsAValidIncomingID(t *testing.T) {
	for _, id := range []string{"abc-123", "4bf92f3577b34da6a3ce929d0e0e4736", strings.Repeat("x", 128), "!~"} {
		seen, rec := serve(id)
		assert.Equal(t, id, seen)
		assert.Equal(t, id, rec.Header().Get(Header))
	}
}

func TestMiddlewareReplacesAnInvalidIncomingID(t *testing.T) {
	for name, id := range map[string]string{
		"too long":       strings.Repeat("x", 129),
		"with a space":   "abc 123",
		"with a control": "abc\x01",
		"not ascii":      "ünïcode",
	} {
		t.Run(name, func(t *testing.T) {
			seen, rec := serve(id)
			assert.Regexp(t, generated, seen)
			assert.Equal(t, seen, rec.Header().Get(Header))
		})
	}
}
