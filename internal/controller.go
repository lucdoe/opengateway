package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceController struct {
	Service IService
}

func SController(service IService) *ServiceController {
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

func (sc *ServiceController) GetServiceByID(c *gin.Context) {
	id := c.Param("id")
	service, err := sc.Service.GetServiceByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, service)
}

func (sc *ServiceController) CreateService(c *gin.Context) {
	var service Service
	if err := c.BindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := sc.Service.CreateService(&service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, service)
}

func (sc *ServiceController) UpdateService(c *gin.Context) {
	var service Service
	if err := c.BindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := sc.Service.UpdateService(&service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, service)
}
