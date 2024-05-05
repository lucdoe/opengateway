package app

import (
	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

func SetupRoutes(r internal.RouterInterface, config *internal.Config) {
	for _, service := range config.Services {
		for _, endpoint := range service.Endpoints {
			hc := HandlerConfig{
				EndpointURL: service.URL + endpoint.Path,
				Endpoint:    endpoint,
				Key:         service.SecretKey,
				CheckKey:    endpoint.Auth.ApplyAuth,
			}

			switch endpoint.HTTPMethod {
			case "GET":
				r.GET(endpoint.Path, hc.AssembleEndpointMiddlewares()...)
			case "POST":
				r.POST(endpoint.Path, hc.AssembleEndpointMiddlewares()...)
			case "PUT":
				r.PUT(endpoint.Path, hc.AssembleEndpointMiddlewares()...)
			case "PATCH":
				r.PATCH(endpoint.Path, hc.AssembleEndpointMiddlewares()...)
			}
		}
	}
}

func APIGatewayApp(router internal.RouterInterface, config *internal.Config) (*internal.App, error) {
	middlewares.InitilizeMiddlewares(router, config)
	SetupRoutes(router, config)
	return &internal.App{Router: router}, nil
}
