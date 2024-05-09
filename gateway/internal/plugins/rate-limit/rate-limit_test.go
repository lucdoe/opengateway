package ratelimit_test

import (
	"errors"
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

func (m *MockRateLimiter) Expire(key string, window time.Duration) error {
	args := m.Called(key, window)
	return args.Error(0)
}

func TestRateLimitService(t *testing.T) {
	mockLimiter := new(MockRateLimiter)
	limit := int64(10)
	window := time.Minute
	service := ratelimit.NewRateLimitService(mockLimiter, limit, window)

	tests := []struct {
		name          string
		key           string
		incrementResp int64
		incrementErr  error
		expectedError string
	}{
		{
			name:          "Below limit",
			key:           "user1",
			incrementResp: 5,
		},
		{
			name:          "At limit",
			key:           "user2",
			incrementResp: 10,
		},
		{
			name:          "Above limit",
			key:           "user3",
			incrementResp: 11,
			expectedError: "rate limit exceeded",
		},
		{
			name:          "Storage error",
			key:           "user4",
			incrementErr:  errors.New("storage failure"),
			expectedError: "storage failure",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockLimiter.On("Increment", tc.key, window).Return(tc.incrementResp, tc.incrementErr).Once()

			count, remaining, _, err := service.RateLimit(tc.key)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.incrementResp, count)
				assert.Equal(t, limit-count, remaining)
			}

			mockLimiter.AssertExpectations(t)
		})
	}
}
