package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type GZIPCompressor interface {
	Compress(next http.Handler) http.Handler
}

type concreteGZIPCompressor struct{}

func NewGZIPCompressor() GZIPCompressor {
	return &concreteGZIPCompressor{}
}

func (c *concreteGZIPCompressor) Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var hasGZIPCompressionHeader bool = strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		if hasGZIPCompressionHeader {
			gw := gzip.NewWriter(w)
			defer gw.Close()

			w.Header().Set("Content-Encoding", "gzip")
			wrappedWriter := &gzipResponseWriter{Writer: gw, ResponseWriter: w}

			next.ServeHTTP(wrappedWriter, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}
