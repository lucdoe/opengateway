package middleware

import (
	"net/http"
	"strings"
)

type CORSMiddleware interface {
	Handle(next http.Handler) http.Handler
}

type concreteCORSMiddleware struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

func NewCORSMiddleware(origins, methods, headers []string) CORSMiddleware {
	return &concreteCORSMiddleware{
		AllowedOrigins: origins,
		AllowedMethods: methods,
		AllowedHeaders: headers,
	}
}

func (c *concreteCORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if len(c.AllowedOrigins) == 0 || contains(c.AllowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if r.Method == http.MethodOptions {
			if len(c.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", joinStrings(c.AllowedMethods, ","))
			}
			if len(c.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(c.AllowedHeaders, ","))
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func joinStrings(slice []string, sep string) string {
	return strings.Join(slice, sep)
}
