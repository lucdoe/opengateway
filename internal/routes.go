package internal

// Routes returns the routes for the gin application.

import (
	"github.com/gin-gonic/gin"
)

func (app *App) SetupRoutes() *gin.Engine {
	r := app.Router

	r.GET("/services", SController(&Service{}).GetAllServices)
	r.GET("/services/:id", SController(&Service{}).GetServiceByID)

	return r
}
