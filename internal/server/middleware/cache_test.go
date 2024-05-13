package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucdoe/open-gateway/gateway/internal"
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

func TestWriteResponse(t *testing.T) {
	w := httptest.NewRecorder()
	content := []byte("Hello, world!")
	statusCode := http.StatusOK
	contentType := "text/plain"

	responseUtil := middleware.StandardResponseUtil{}
	responseUtil.WriteResponse(w, statusCode, contentType, content)

	res := w.Result()
	resBody, _ := io.ReadAll(res.Body)

	if res.StatusCode != statusCode {
		t.Errorf("Expected status code %d, got %d", statusCode, res.StatusCode)
	}
	if res.Header.Get("Content-Type") != contentType {
		t.Errorf("Expected content type %s, got %s", contentType, res.Header.Get("Content-Type"))
	}
	if string(resBody) != string(content) {
		t.Errorf("Expected response body %s, got %s", string(content), string(resBody))
	}
}

type ResponseRecorderStub struct {
	httptest.ResponseRecorder
	StatusCode int
}

func (rs *ResponseRecorderStub) WriteHeader(statusCode int) {
	rs.StatusCode = statusCode
	rs.ResponseRecorder.WriteHeader(statusCode)
}

func TestCopyStatusAndHeader(t *testing.T) {
	src := internal.NewResponseRecorder(httptest.NewRecorder())
	dst := httptest.NewRecorder()

	src.Header().Set("Content-Type", "application/json")
	src.WriteHeader(http.StatusNotFound)

	responseUtil := middleware.StandardResponseUtil{}
	responseUtil.CopyStatusAndHeader(src, dst)

	if dst.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", dst.Header().Get("Content-Type"))
	}

	if dst.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, dst.Code)
	}
}
