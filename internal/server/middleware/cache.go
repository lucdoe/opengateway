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
	"net/http"
	"time"

	"github.com/lucdoe/opengateway/internal"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
)

type CacheMiddleware struct {
	Cache        cache.Cache
	ResponseUtil ResponseUtil
}

func NewCacheMiddleware(c cache.Cache, ru ResponseUtil) *CacheMiddleware {
	return &CacheMiddleware{
		Cache:        c,
		ResponseUtil: ru,
	}
}

type ResponseUtil interface {
	CopyStatusAndHeader(src, dst http.ResponseWriter)
	WriteResponse(w http.ResponseWriter, statusCode int, contentType string, content []byte)
}

type StandardResponseUtil struct{}

func (s *StandardResponseUtil) CopyStatusAndHeader(src, dst http.ResponseWriter) {
	dst.WriteHeader(src.(*internal.ResponseRecorder).StatusCode)
	for k, vv := range src.Header() {
		for _, v := range vv {
			dst.Header().Add(k, v)
		}
	}
}

func (s *StandardResponseUtil) WriteResponse(w http.ResponseWriter, statusCode int, contentType string, content []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	w.Write(content)
}

func (cm *CacheMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheKey := cm.Cache.GenerateCacheKey(r)
		cachedResponse, err := cm.Cache.Get(cacheKey)

		if handleCacheHit(w, cachedResponse, err) {
			return
		}

		recorder := internal.NewResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if recorder.StatusCode == http.StatusOK {
			handleCacheMiss(cm, cacheKey, recorder)
		}

		copyResponseHeaders(w, recorder)
	})
}

func handleCacheHit(w http.ResponseWriter, cachedResponse string, err error) bool {
	if err == nil && cachedResponse != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cachedResponse))
		return true
	}
	return false
}

func handleCacheMiss(cm *CacheMiddleware, cacheKey string, recorder *internal.ResponseRecorder) {
	responseBody := recorder.Body.String()
	cm.Cache.Set(cacheKey, responseBody, 10*time.Minute)
}

func copyResponseHeaders(w http.ResponseWriter, recorder *internal.ResponseRecorder) {
	for k, vv := range recorder.Header() {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
}
