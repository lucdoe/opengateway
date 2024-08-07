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

package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal"
	"github.com/stretchr/testify/mock"
)

const (
	contentTypeHeader = "Content-Type"
	JSONContentType   = "application/json"
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
	middlewareInstance := NewCacheMiddleware(mockCache, mockResponseUtil)

	cacheKey := "cacheKey"
	responseBody := `{"status":"ok"}`
	mockCache.On("GenerateCacheKey", mock.Anything).Return(cacheKey)
	mockCache.On("Get", cacheKey).Return("", nil)
	mockCache.On("Set", cacheKey, responseBody, 10*time.Minute).Return(nil)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	})

	middlewareInstance.Middleware(testHandler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != responseBody {
		t.Errorf("Expected response body %s, got %s", responseBody, recorder.Body.String())
	}
	mockCache.AssertExpectations(t)
}

func TestWriteResponse(t *testing.T) {
	w := httptest.NewRecorder()
	content := []byte("Hello, world!")
	statusCode := http.StatusOK
	contentType := "text/plain"

	responseUtil := StandardResponseUtil{}
	responseUtil.WriteResponse(w, statusCode, contentType, content)

	res := w.Result()
	resBody, _ := io.ReadAll(res.Body)

	if res.StatusCode != statusCode {
		t.Errorf("Expected status code %d, got %d", statusCode, res.StatusCode)
	}
	if res.Header.Get(contentTypeHeader) != contentType {
		t.Errorf("Expected content type %s, got %s", contentType, res.Header.Get(contentTypeHeader))
	}
	if string(resBody) != string(content) {
		t.Errorf("Expected response body %s, got %s", string(content), string(resBody))
	}
}

func TestCopyStatusAndHeader(t *testing.T) {
	src := internal.NewResponseRecorder(httptest.NewRecorder())
	dst := httptest.NewRecorder()

	src.Header().Set(contentTypeHeader, JSONContentType)
	src.WriteHeader(http.StatusNotFound)

	responseUtil := StandardResponseUtil{}
	responseUtil.CopyStatusAndHeader(src, dst)

	if dst.Header().Get(contentTypeHeader) != JSONContentType {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", dst.Header().Get(contentTypeHeader))
	}

	if dst.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, dst.Code)
	}
}

func TestHandleCacheHit(t *testing.T) {
	recorder := httptest.NewRecorder()
	cachedResponse := `{"status":"ok"}`

	if !handleCacheHit(recorder, cachedResponse, nil) {
		t.Errorf("Expected cache hit to be handled")
	}

	res := recorder.Result()
	resBody, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}
	if res.Header.Get(contentTypeHeader) != JSONContentType {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", res.Header.Get(contentTypeHeader))
	}
	if string(resBody) != cachedResponse {
		t.Errorf("Expected response body %s, got %s", cachedResponse, string(resBody))
	}
}

func TestHandleCacheMiss(t *testing.T) {
	mockCache := new(MockCache)
	recorder := internal.NewResponseRecorder(httptest.NewRecorder())
	responseBody := `{"status":"ok"}`
	recorder.WriteHeader(http.StatusOK)
	recorder.Write([]byte(responseBody))

	cacheKey := "cacheKey"
	mockCache.On("Set", cacheKey, responseBody, 10*time.Minute).Return(nil)

	cm := &CacheMiddleware{Cache: mockCache}
	handleCacheMiss(cm, cacheKey, recorder)

	mockCache.AssertCalled(t, "Set", cacheKey, responseBody, 10*time.Minute)
}
