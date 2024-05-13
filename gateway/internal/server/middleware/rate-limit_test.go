package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ratelimit "github.com/lucdoe/open-gateway/gateway/internal/plugins/rate-limit"
	"github.com/lucdoe/open-gateway/gateway/internal/server/middleware"
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
	middleware := middleware.NewRateLimitMiddleware(mockRateLimiter)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("request allowed", func(t *testing.T) {
		mockRateLimiter.On("RateLimit", mock.Anything).Return(int64(1), int64(9), time.Minute, nil)
		mockRateLimiter.On("GetLimit").Return(int64(10))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/", nil)

		middleware.Middleware(testHandler).ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "10", recorder.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "9", recorder.Header().Get("X-RateLimit-Remaining"))
	})

}

func TestRateLimitMiddlewareRateLimitExceeded(t *testing.T) {
	mockRateLimiter := new(MockRateLimiter)
	middleware := middleware.NewRateLimitMiddleware(mockRateLimiter)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		mockRateLimiter.On("RateLimit", mock.Anything).Return(int64(0), int64(0), time.Minute, ratelimit.ErrRateLimitExceeded)
		mockRateLimiter.On("GetLimit").Return(int64(10))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/", nil)

		middleware.Middleware(testHandler).ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		assert.Equal(t, "10", recorder.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", recorder.Header().Get("X-RateLimit-Remaining"))
	})

}
