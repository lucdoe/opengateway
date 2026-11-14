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

package ratelimit_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	ratelimit "github.com/lucdoe/opengateway/internal/plugins/rate-limit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRateLimiter struct {
	mock.Mock
	ratelimit.RateLimiter
}

func (m *MockRateLimiter) RateLimit(key string) (int64, int64, time.Duration, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Get(1).(int64), args.Get(2).(time.Duration), args.Error(3)
}

func (m *MockRateLimiter) GetLimit() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func TestRateLimitMiddlewareAllowed(t *testing.T) {
	mockRateLimiter := new(MockRateLimiter)
	middleware := ratelimit.NewMiddleware(mockRateLimiter)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("request allowed", func(t *testing.T) {
		mockRateLimiter.On("RateLimit", mock.Anything).Return(int64(1), int64(9), time.Minute, nil)
		mockRateLimiter.On("GetLimit").Return(int64(10))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/", nil)

		middleware(testHandler).ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "10", recorder.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "9", recorder.Header().Get("X-RateLimit-Remaining"))
		assert.Equal(t, "60", recorder.Header().Get("X-RateLimit-Reset"), "seconds until the reset")
		assert.Empty(t, recorder.Header().Get("Retry-After"))
	})

}

func TestRateLimitMiddlewareRateLimitExceeded(t *testing.T) {
	mockRateLimiter := new(MockRateLimiter)
	middleware := ratelimit.NewMiddleware(mockRateLimiter)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		mockRateLimiter.On("RateLimit", mock.Anything).Return(int64(11), int64(0), 1500*time.Millisecond, ratelimit.ErrRateLimitExceeded)
		mockRateLimiter.On("GetLimit").Return(int64(10))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/", nil)

		middleware(testHandler).ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		assert.Equal(t, "10", recorder.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", recorder.Header().Get("X-RateLimit-Remaining"))
		assert.Equal(t, "2", recorder.Header().Get("X-RateLimit-Reset"), "rounded up to whole seconds")
		assert.Equal(t, "2", recorder.Header().Get("Retry-After"))
		assert.Equal(t, "Rate limit exceeded\n", recorder.Body.String())
	})

}

func TestRateLimitMiddlewareCountsPerClientIP(t *testing.T) {
	tests := []struct {
		name        string
		remoteAddrs []string
		expected    []string
	}{
		{
			name:        "connections from one client share a key",
			remoteAddrs: []string{"203.0.113.5:51000", "203.0.113.5:51001"},
			expected:    []string{"ratelimit:203.0.113.5", "ratelimit:203.0.113.5"},
		},
		{
			name:        "different clients get different keys",
			remoteAddrs: []string{"203.0.113.5:51000", "[2001:db8::1]:51000"},
			expected:    []string{"ratelimit:203.0.113.5", "ratelimit:2001:db8::1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var keys []string
			mockRateLimiter := new(MockRateLimiter)
			mockRateLimiter.On("RateLimit", mock.Anything).
				Run(func(args mock.Arguments) { keys = append(keys, args.String(0)) }).
				Return(int64(1), int64(9), time.Minute, nil)
			mockRateLimiter.On("GetLimit").Return(int64(10))
			handler := ratelimit.NewMiddleware(mockRateLimiter)(http.NotFoundHandler())

			for _, addr := range tc.remoteAddrs {
				request := httptest.NewRequest("GET", "/", nil)
				request.RemoteAddr = addr
				request.Header.Set("X-Forwarded-For", "10.0.0.1")
				handler.ServeHTTP(httptest.NewRecorder(), request)
			}

			assert.Equal(t, tc.expected, keys)
		})
	}
}

func TestRateLimitMiddlewareRejectsWrappedLimitErrors(t *testing.T) {
	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true })
	mockRateLimiter := new(MockRateLimiter)
	mockRateLimiter.On("RateLimit", mock.Anything).Return(int64(11), int64(0), time.Second, fmt.Errorf("checking limit: %w", ratelimit.ErrRateLimitExceeded))
	mockRateLimiter.On("GetLimit").Return(int64(10))

	recorder := httptest.NewRecorder()
	ratelimit.NewMiddleware(mockRateLimiter)(next).ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.False(t, reached, "the request must not be forwarded")
}

// A store outage must not take down the services behind the plugin.
func TestRateLimitMiddlewareLetsRequestsThroughWhenTheStoreFails(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	reached := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached++ })
	mockRateLimiter := new(MockRateLimiter)
	// Same text as the limit error, but not that error.
	mockRateLimiter.On("RateLimit", mock.Anything).Return(int64(0), int64(0), time.Duration(0), errors.New("rate limit exceeded"))
	handler := ratelimit.NewMiddleware(mockRateLimiter)(next)

	var last *httptest.ResponseRecorder
	for i := 0; i < 3; i++ {
		last = httptest.NewRecorder()
		handler.ServeHTTP(last, httptest.NewRequest("GET", "/", nil))
	}

	assert.Equal(t, 3, reached, "every request is forwarded")
	assert.Equal(t, http.StatusOK, last.Code)
	assert.Empty(t, last.Header().Get("X-RateLimit-Limit"), "no limit is claimed that was not checked")
	assert.Equal(t, 1, strings.Count(logs.String(), "rate-limit: store failed"), "the outage is logged once, not per request")
	assert.Contains(t, logs.String(), "rate limit exceeded")
}
