package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

func handleJSONValidation(body internal.BodyField) gin.HandlerFunc {
	return func(c *gin.Context) {
		if body.ApplyValidation {
			middlewares.ValidateJSONFields(body)
		}
		c.Next()
	}
}

func SetupRoutes(r internal.RouterInterface, config *internal.Config) {
	for serviceName, service := range config.Services {
		URL := fmt.Sprintf("%s:%d", service.URL, service.PORT)

		for _, endpoint := range service.Endpoints {
			endpointURL := URL + endpoint.Path
			switch endpoint.HTTPMethod {
			case "GET":
				r.GET(endpoint.Path, handleJSONValidation(endpoint.Body))
			case "POST":
				r.POST(endpoint.Path, handleJSONValidation(endpoint.Body))
			case "PUT":
				r.PUT(endpoint.Path, handleJSONValidation(endpoint.Body))
			case "PATCH":
				r.PATCH(endpoint.Path, handleJSONValidation(endpoint.Body))
			}

			r.Use(middlewares.Proxy(serviceName, endpointURL))
		}
	}
}

func APIGatewayApp(router internal.RouterInterface, config *internal.Config) (*internal.App, error) {
	SetupRoutes(router, config)
	return &internal.App{Router: router}, nil
}
