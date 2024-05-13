package middleware

import (
	"net/http"
	"time"

	"github.com/lucdoe/open-gateway/gateway/internal"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
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
		if err == nil && cachedResponse != "" {
			cm.ResponseUtil.WriteResponse(w, http.StatusOK, "application/json", []byte(cachedResponse))
			return
		}

		recorder := internal.NewResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if recorder.StatusCode == http.StatusOK {
			responseBody := recorder.Body.String()
			cm.Cache.Set(cacheKey, responseBody, 10*time.Minute)
			cm.ResponseUtil.WriteResponse(w, http.StatusOK, "application/json", []byte(responseBody))
		} else {
			cm.ResponseUtil.CopyStatusAndHeader(recorder, w)
		}
	})
}
