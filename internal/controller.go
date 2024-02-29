package internal

import (
	"fmt"
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
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	handleSuccess(c, http.StatusOK, services)
}

func (sc *ServiceController) GetServiceByID(c *gin.Context) {
	id := c.Param("id")
	service, err := sc.Service.GetServiceByID(id)

	if err != nil {
		handleError(c, http.StatusNotFound, err)
		return
	}

	handleSuccess(c, http.StatusOK, service)
}

func (sc *ServiceController) CreateService(c *gin.Context) {
	var service Service

	if err := c.BindJSON(&service); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := sc.Service.CreateService(&service); err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	setHeaders(c, map[string]string{"Location": fmt.Sprintf("/services/%d", service.ID)})
	handleSuccess(c, http.StatusCreated, service.ID)
}

func (sc *ServiceController) UpdateService(c *gin.Context) {
	var service Service

	if err := c.BindJSON(&service); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := sc.Service.UpdateService(&service); err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	setHeaders(c, map[string]string{"Location": fmt.Sprintf("/services/%d", service.ID)})
	handleSuccess(c, http.StatusOK, service.ID)
}
