package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucdoe/open-gateway/gateway/internal/server/middleware"
	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Increment(key string, window time.Duration) (int64, error) {
	args := m.Called(key, window)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(key string, value string, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) GenerateCacheKey(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

type MockResponseUtil struct {
	mock.Mock
}

func (m *MockResponseUtil) CopyStatusAndHeader(src, dst http.ResponseWriter) {
	m.Called(src, dst)
}

func (m *MockResponseUtil) WriteResponse(w http.ResponseWriter, statusCode int, contentType string, content []byte) {
	m.Called(w, statusCode, contentType, content)
}

func TestCacheMiddleware(t *testing.T) {
	mockCache := new(MockCache)
	mockResponseUtil := new(MockResponseUtil)
	middlewareInstance := middleware.NewCacheMiddleware(mockCache, mockResponseUtil)

	cacheKey := "cacheKey"
	responseBody := `{"status":"ok"}`
	mockCache.On("GenerateCacheKey", mock.Anything).Return(cacheKey)
	mockCache.On("Get", cacheKey).Return("", nil)
	mockCache.On("Set", cacheKey, responseBody, 10*time.Minute).Return(nil)

	mockResponseUtil.On("WriteResponse", mock.Anything, http.StatusOK, "application/json", []byte(responseBody)).Once()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	})

	middlewareInstance.Middleware(testHandler).ServeHTTP(recorder, request)

	mockResponseUtil.AssertCalled(t, "WriteResponse", mock.Anything, http.StatusOK, "application/json", []byte(responseBody))
	mockCache.AssertExpectations(t)
}
