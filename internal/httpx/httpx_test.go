// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package httpx

import (
	"net/http/httptest"
	"testing"
	"time"

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

			assert.Equal(t, tc.expected, ClientIP(req))
		})
	}
}

func TestSeconds(t *testing.T) {
	assert.Equal(t, "0", Seconds(0))
	assert.Equal(t, "1", Seconds(time.Millisecond))
	assert.Equal(t, "2", Seconds(1500*time.Millisecond))
	assert.Equal(t, "60", Seconds(time.Minute))
}
