package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

func BodySanitize(c *gin.Context) {
	p := bluemonday.UGCPolicy()

	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(p.SanitizeBytes(b)))

	c.Next()
}
