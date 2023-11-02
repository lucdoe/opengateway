package middlewares

import "github.com/gin-gonic/gin"

func InitilizeMiddlewares(r *gin.Engine) {
	r.Use(gin.Recovery())

	r.Use(LogRequest)
	r.Use(BodyLimit)
	r.Use(SecurityHeaders)
	r.Use(RateLimit)
	r.Use(CORS)
}
