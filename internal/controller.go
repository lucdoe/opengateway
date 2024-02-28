package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceController struct {
	Service IService
}

func NewServiceController(service IService) *ServiceController {
	return &ServiceController{
		Service: service,
	}
}

func (sc *ServiceController) GetAllServices(c *gin.Context) {
	services, err := sc.Service.GetAllServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services)
}
