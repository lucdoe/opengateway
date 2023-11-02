package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LimitBodySize(c *gin.Context) {
	var b []byte

	if c.Request.Body != nil {
		b, _ = io.ReadAll(c.Request.Body)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(b))

	size := int64(len(b))

	if size > 1<<20 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too big"})
		c.Abort()
		return
	}

	c.Next()
}
