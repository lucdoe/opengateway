package middlewares

import (
	"github.com/gin-gonic/gin"
)

func Proxy(serviceName string, URL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.URL.Host = serviceName
		c.Request.URL.Scheme = "http"
		c.Request.URL.Path = URL
	}
}
