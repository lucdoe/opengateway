package middleware

import (
	"net/http"
	"time"

	"github.com/lucdoe/open-gateway/gateway/internal"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cache"
)

type CacheMiddleware struct {
	Cache cache.Cache
}

func NewCacheMiddleware(c cache.Cache) *CacheMiddleware {
	return &CacheMiddleware{Cache: c}
}

func (cm *CacheMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheKey := cm.Cache.GenerateCacheKey(r)

		cachedResponse, err := cm.Cache.Get(cacheKey)
		if err == nil && cachedResponse != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cachedResponse))
			return
		}

		recorder := internal.NewResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if recorder.StatusCode == http.StatusOK {
			responseBody := recorder.Body.String()
			cm.Cache.Set(cacheKey, responseBody, 10*time.Minute)
			w.Write([]byte(responseBody))
		} else {
			recorder.CopyBody(w)
		}
	})
}
