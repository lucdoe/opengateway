// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		remoteAddr string
		expected   string
	}{
		{"203.0.113.5:51000", "203.0.113.5"},
		{"[2001:db8::1]:51000", "2001:db8::1"},
		{"203.0.113.5", "203.0.113.5"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.remoteAddr, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remoteAddr
			req.Header.Set("X-Forwarded-For", "10.0.0.1")

			assert.Equal(t, tc.expected, clientIP(req))
		})
	}
}
