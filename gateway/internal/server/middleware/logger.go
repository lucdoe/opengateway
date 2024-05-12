package middleware

import (
	"fmt"
	"net/http"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
)

type LoggingMiddleware struct {
	Logger logger.Logger
}

func NewLoggingMiddleware(l logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		Logger: l,
	}
}

func (lm *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		details := fmt.Sprintf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		lm.Logger.Info("Request", details)
		next.ServeHTTP(w, r)
	})
}
