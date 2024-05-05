package app

import (
	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

type GinRouter struct {
	*gin.Engine
}

func (gr GinRouter) GET(path string, handlers ...gin.HandlerFunc) {
	gr.Engine.GET(path, handlers...)
}
func (gr GinRouter) POST(path string, handlers ...gin.HandlerFunc) {
	gr.Engine.POST(path, handlers...)
}
func (gr GinRouter) PATCH(path string, handlers ...gin.HandlerFunc) {
	gr.Engine.PATCH(path, handlers...)
}
func (gr GinRouter) PUT(path string, handlers ...gin.HandlerFunc) {
	gr.Engine.PUT(path, handlers...)
}
func (gr GinRouter) OPTIONS(path string, handlers ...gin.HandlerFunc) {
	gr.Engine.OPTIONS(path, handlers...)
}

func (gr GinRouter) Use(middlewares ...gin.HandlerFunc) {
	gr.Engine.Use(middlewares...)
}
func (gr GinRouter) Run(addr ...string) error {
	return gr.Engine.Run(addr...)
}

func InitializeRouter() GinRouter {
	ginRouter := gin.New()
	return GinRouter{Engine: ginRouter}
}

type HandlerConfig struct {
	EndpointURL string
	Endpoint    internal.Endpoint
	Key         string
	CheckKey    bool
}

func (hc HandlerConfig) AssembleEndpointMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middlewares.ValidateToken(hc.CheckKey, hc.Key),
		middlewares.ValidateJSONFields(hc.Endpoint.Body, hc.Endpoint.Body.ApplyValidation),
		middlewares.ReverseProxy(hc.EndpointURL),
	}
}
