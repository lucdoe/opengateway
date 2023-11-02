package middlewares

import "github.com/gin-gonic/gin"

func InitilizeMiddlewares(r *gin.Engine) {
	r.Use(gin.Recovery())

	// add all middlewares here
	r.Use(LogRequest)
	r.Use(SecurityHeaders)
	r.Use(CORS)
}
