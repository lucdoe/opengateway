// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package requestid gives every request an ID that the upstream, the log
// and the client all see, so one request can be followed across services.
package requestid

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// Header carries the request ID, to the upstream and back to the client.
const Header = "X-Request-ID"

// maxLength is the longest ID taken from a client. Longer ones are replaced.
const maxLength = 128

// Middleware keeps the ID a client or a proxy in front sent, when it is
// 1 to 128 visible ASCII characters, and otherwise sets a new random one. The
// ID is set on the request, so the upstream and later plugins read it, and on
// the response, including responses the gateway answers itself.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(Header)
		if !valid(id) {
			id = newID()
		}
		r.Header.Set(Header, id)
		w.Header().Set(Header, id)
		next.ServeHTTP(w, r)
	})
}

// valid reports whether id is safe to pass on and log as it is.
func valid(id string) bool {
	if id == "" || len(id) > maxLength {
		return false
	}
	for i := 0; i < len(id); i++ {
		if id[i] < '!' || id[i] > '~' {
			return false
		}
	}
	return true
}

// newID returns 128 random bits as 32 hex characters.
func newID() string {
	var b [16]byte
	rand.Read(b[:]) // fails only when the OS has no source of randomness
	return hex.EncodeToString(b[:])
}
