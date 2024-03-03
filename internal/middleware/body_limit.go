package middleware

import (
	"io"
	"net/http"
)

func BodyLimitHandler(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)

			buf := make([]byte, limit+1)
			n, err := r.Body.Read(buf)
			if err != nil && err != io.EOF {
				http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			if int64(n) == limit+1 {
				http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
