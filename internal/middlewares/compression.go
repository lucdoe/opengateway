package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type GZIPResponseWriter struct {
	io.Writer
	gin.ResponseWriter
}

func GZIPCompression() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {

			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			c.Writer = &GZIPResponseWriter{Writer: gz, ResponseWriter: c.Writer}

			c.Header("Content-Encoding", "gzip")
		}

		c.Next()
	}
}

func (g *GZIPResponseWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}
