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
	"errors"
	"testing"
	"time"

	ratelimit "github.com/lucdoe/opengateway/internal/plugins/rate-limit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testIP = "123.123.123.123"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Increment(key string, window time.Duration) (int64, error) {
	args := m.Called(key, window)
	return args.Get(0).(int64), args.Error(1)
}

func TestRateLimitService(t *testing.T) {
	rlConfig := ratelimit.RateLimitConfig{
		Store:  new(MockCache),
		Limit:  10,
		Window: time.Minute,
	}

	service := ratelimit.NewRateLimitService(rlConfig)

	mockCache := service.Store.(*MockCache)

	window := time.Minute
	limit := int64(10)

	t.Run("Under limit", func(t *testing.T) {
		mockCache.On("Increment", testIP, window).Return(int64(1), nil).Once()
		count, remaining, _, err := service.RateLimit(testIP)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, limit-1, remaining)
		mockCache.AssertExpectations(t)
	})

	t.Run("Rate limit exceeded", func(t *testing.T) {
		mockCache.On("Increment", testIP, window).Return(int64(11), nil).Once()
		count, remaining, _, err := service.RateLimit(testIP)
		assert.Error(t, err)
		assert.Equal(t, "rate limit exceeded", err.Error())
		assert.Equal(t, int64(11), count)
		assert.Equal(t, int64(0), remaining)
		mockCache.AssertExpectations(t)
	})

	t.Run("Increment error", func(t *testing.T) {
		mockCache.On("Increment", testIP, window).Return(int64(0), errors.New("redis error")).Once()
		_, _, _, err := service.RateLimit(testIP)
		assert.Error(t, err)
		assert.Equal(t, "redis error", err.Error())
		mockCache.AssertExpectations(t)
	})

	t.Run("Check rate limit", func(t *testing.T) {
		assert.Equal(t, int64(10), service.GetLimit())
	})
}

func TestRateLimitConfigInitialization(t *testing.T) {
	mockCache := new(MockCache)
	cfg := ratelimit.RateLimitConfig{
		Store:  mockCache,
		Limit:  20,
		Window: 2 * time.Minute,
	}

	service := ratelimit.NewRateLimitService(cfg)
	assert.NotNil(t, service)
	assert.Equal(t, int64(20), service.GetLimit())
	assert.Equal(t, 2*time.Minute, service.Window)
	assert.Equal(t, mockCache, service.Store)
}
