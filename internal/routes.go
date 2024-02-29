package internal

import (
	"github.com/gin-gonic/gin"
)

const servicesPath = "/services"

func (app *App) SetupRoutes() *gin.Engine {
	r := app.Router

	// services routes
	r.GET(servicesPath, SController(&Service{}).GetAllServices)
	r.GET(servicesPath+"/:id", SController(&Service{}).GetServiceByID)
	r.POST(servicesPath, SController(&Service{}).CreateService)
	r.PUT(servicesPath+"/:id", SController(&Service{}).UpdateService)

	return r
}
