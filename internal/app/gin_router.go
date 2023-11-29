package app

import (
	"github.com/gin-gonic/gin"
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
