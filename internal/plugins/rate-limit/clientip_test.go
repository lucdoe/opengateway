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
