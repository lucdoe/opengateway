package app

import (
	"fmt"

	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

func setupRoutes(r internal.RouterInterface, config *internal.Config) {
	for serviceName, service := range config.Services {
		URL := fmt.Sprintf("%s:%d", service.URL, service.PORT)

		for _, endpoint := range service.Endpoints {
			endpointURL := URL + endpoint.Path
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

			r.Use(middlewares.Proxy(serviceName, endpointURL))
		}
	}
}

func APIGatewayApp(router internal.RouterInterface, config *internal.Config) (*internal.App, error) {
	setupRoutes(router, config)
	return &internal.App{Router: router}, nil
}
