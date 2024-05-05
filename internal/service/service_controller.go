package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/gateway_admin_api/internal/databases"
	"github.com/lucdoe/gateway_admin_api/internal/helpers"
)

type ServiceController struct {
	Service IService
}

func ServiceControllerConstructor(service IService) *ServiceController {
	return &ServiceController{
		Service: service,
	}
}

func (sc *ServiceController) AllServicesHandler(c *gin.Context) {
	services, err := sc.Service.GetAll()

	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	helpers.HandleSuccess(c, http.StatusOK, services)
}

func (sc *ServiceController) ServiceByIDHandler(c *gin.Context) {
	id := c.Param("id")
	service, err := sc.Service.GetByID(id)

	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, err)
		return
	}

	helpers.HandleSuccess(c, http.StatusOK, service)
}

func (sc *ServiceController) CreateServiceHandler(c *gin.Context) {
	var service databases.Service

	if err := c.BindJSON(&service); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, err)
		return
	}
	if err := sc.Service.Create(&service); err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	helpers.SetHeaders(c, map[string]string{"Location": fmt.Sprintf("/services/%d", service.ID)})
	helpers.HandleSuccess(c, http.StatusCreated, gin.H{"ID": service.ID})
}

func (sc *ServiceController) UpdateServiceHandler(c *gin.Context) {
	var service databases.Service

	if err := c.BindJSON(&service); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, err)
		return
	}
	if err := sc.Service.Update(&service); err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	helpers.SetHeaders(c, map[string]string{"Location": fmt.Sprintf("/services/%d", service.ID)})
	helpers.HandleSuccess(c, http.StatusOK, gin.H{"ID": service.ID})
}
