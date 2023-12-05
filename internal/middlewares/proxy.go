package middlewares

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func Proxy(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		target, err := url.Parse(targetURL)
		if err != nil {
			log.Printf("Error parsing target URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		log.Printf("Proxying to URL: %v", target)

		reverseProxy := httputil.NewSingleHostReverseProxy(target)
		reverseProxy.Director = func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path // Adjust if necessary based on your routing logic
			req.Header["X-Forwarded-Host"] = []string{c.Request.Host}
		}

		reverseProxy.ServeHTTP(c.Writer, c.Request)
	}
}
