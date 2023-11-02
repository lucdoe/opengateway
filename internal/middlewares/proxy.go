package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
)

func Proxy(serviceName string, endpoint internal.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.URL.Host = serviceName
		c.Request.URL.Scheme = "http"
		c.Request.URL.Path = endpoint.Path
	}
}
