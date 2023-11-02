package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

type App struct {
	Router *gin.Engine
}

func APIGatewayAPP(config *internal.Config) (*App, error) {
	r := gin.New()

	middlewares.InitilizeMiddlewares(r)

	for serviceName, service := range config.Services {
		for _, endpoint := range service.Endpoints {
			switch endpoint.HTTPMethod {
			case "GET":
				r.GET(endpoint.Path, middlewares.ValidateJSONFields(endpoint.AllowedJSON))
			case "POST":
				r.POST(endpoint.Path, middlewares.ValidateJSONFields(endpoint.AllowedJSON))
			case "PUT":
				r.PUT(endpoint.Path, middlewares.ValidateJSONFields(endpoint.AllowedJSON))
			case "PATCH":
				r.PATCH(endpoint.Path, middlewares.ValidateJSONFields(endpoint.AllowedJSON))
			}
			fmt.Print(serviceName)
		}
	}

	return &App{Router: r}, nil
}
