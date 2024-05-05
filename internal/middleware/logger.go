package middleware

import (
	"net/http"

	"github.com/lucdoe/opengateway/internal/logger"
)

func LoggingMiddleware(logger logger.StandardLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Log("Received request: " + r.Method + " " + r.URL.Path + " " + r.RemoteAddr + " " + r.UserAgent())

			next.ServeHTTP(w, r)
		})
	}
}
