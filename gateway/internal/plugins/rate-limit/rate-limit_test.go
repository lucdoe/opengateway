package ratelimit_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ratelimit "github.com/lucdoe/open-gateway/gateway/internal/plugins/rate-limit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) Increment(key string, window time.Duration) (int64, error) {
	args := m.Called(key, window)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count, err := m.Increment(r.RemoteAddr, 1*time.Minute)
		if err != nil {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
		w.Header().Set("X-Rate-Limit-Count", string(rune(count)))
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	mockLimiter := new(MockRateLimiter)
	limit := int64(10)
	window := time.Minute
	service := ratelimit.NewRateLimitService(mockLimiter, limit, window)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := service.Middleware(handler)

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "123.123.123.123"

	t.Run("Under limit", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		mockLimiter.On("Increment", req.RemoteAddr, window).Return(int64(1), nil).Once()
		middleware.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "10", recorder.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "9", recorder.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("Rate limit exceeded", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		mockLimiter.On("Increment", req.RemoteAddr, window).Return(int64(11), nil).Once()
		middleware.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		assert.Equal(t, "0", recorder.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("Increment error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		mockLimiter.On("Increment", req.RemoteAddr, window).Return(int64(2), errors.New("redis error")).Once()
		middleware.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}
