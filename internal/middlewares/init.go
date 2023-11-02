package middlewares

import "github.com/gin-gonic/gin"

func InitilizeMiddlewares(r *gin.Engine) {
	r.Use(gin.Recovery())

	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Use(LogRequest)
	r.Use(BodyLimit)
	r.Use(BodySanitize)
	r.Use(SecurityHeaders)
	r.Use(RateLimit)
	r.Use(CORS)
}
