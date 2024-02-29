package internal

import (
	"github.com/gin-gonic/gin"
)

const ServicesPath = "/services"

func SetupServiceRoutes(r *gin.Engine) *gin.Engine {

	r.GET(ServicesPath, SController(&Service{}).GetAllServices)
	r.GET(ServicesPath+"/:id", SController(&Service{}).GetServiceByID)
	r.POST(ServicesPath, SController(&Service{}).CreateService)
	r.PUT(ServicesPath+"/:id", SController(&Service{}).UpdateService)

	return r
}
