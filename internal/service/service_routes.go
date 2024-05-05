package service

import (
	"github.com/gin-gonic/gin"
)

const ServicesPath = "/services"

func SetupServiceRoutes(r *gin.Engine, c *ServiceController) *gin.Engine {
	r.GET(ServicesPath, c.AllServicesHandler)
	r.GET(ServicesPath+"/:id", c.ServiceByIDHandler)
	r.POST(ServicesPath, c.CreateServiceHandler)
	r.PUT(ServicesPath+"/:id", c.UpdateServiceHandler)

	return r
}
