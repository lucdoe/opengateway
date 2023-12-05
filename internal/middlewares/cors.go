package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
)

const MaxAgeCORS = 12 * time.Hour

func CORS(config *internal.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.CORS.Apply {
			origins := strings.Join(config.CORS.Origins, ", ")
			headers := strings.Join(config.CORS.Headers, ", ")

			c.Writer.Header().Set("Access-Control-Allow-Origin", origins)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
			c.Writer.Header().Set("Max-Age", fmt.Sprintf("%f", MaxAgeCORS.Seconds()))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
