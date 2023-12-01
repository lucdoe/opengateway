package app

import (
	"fmt"

	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

func SetupRoutes(r internal.RouterInterface, config *internal.Config) {
	for serviceName, service := range config.Services {
		URL := fmt.Sprintf("%s:%d", service.URL, service.PORT)

		for _, endpoint := range service.Endpoints {
			endpointURL := URL + endpoint.Path
			switch endpoint.HTTPMethod {
			case "GET":
				r.GET(endpoint.Path, middlewares.ValidateJSONFields(endpoint.Body))
			case "POST":
				r.POST(endpoint.Path, middlewares.ValidateJSONFields(endpoint.Body))
			case "PUT":
				r.PUT(endpoint.Path, middlewares.ValidateJSONFields(endpoint.Body))
			case "PATCH":
				r.PATCH(endpoint.Path, middlewares.ValidateJSONFields(endpoint.Body))
			}

			r.Use(middlewares.Proxy(serviceName, endpointURL))
		}
	}
}

func APIGatewayApp(router internal.RouterInterface, config *internal.Config) (*internal.App, error) {
	SetupRoutes(router, config)
	return &internal.App{Router: router}, nil
}
