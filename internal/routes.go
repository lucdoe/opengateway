package internal

import (
	"github.com/gin-gonic/gin"
)

func (app *App) SetupRoutes() *gin.Engine {
	r := app.Router

	// services routes
	r.GET("/services", SController(&Service{}).GetAllServices)
	r.GET("/services/:id", SController(&Service{}).GetServiceByID)
	r.POST("/services", SController(&Service{}).CreateService)

	return r
}
