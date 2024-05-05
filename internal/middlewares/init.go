package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
)

func InitilizeMiddlewares(r internal.RouterInterface, config *internal.Config) {
	r.Use(gin.Recovery())

	r.Use(GZIPCompression())
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Use(LogRequest)
	r.Use(BodyLimit)
	r.Use(SecurityHeaders)
	r.Use(RateLimit)
	r.Use(CORS(config))
}
