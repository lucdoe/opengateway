package app

import (
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

func proxyRequest(URL string) gin.HandlerFunc {
	return middlewares.Proxy(URL)
}

func attachQueryParams(params []internal.QueryParam) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, param := range params {
			value := c.Query(param.Key)
			c.Set(param.Key, value)
		}
		c.Next()
	}
}

func SetupRoutes(r internal.RouterInterface, config *internal.Config) {
	for _, service := range config.Services {
		for _, endpoint := range service.Endpoints {
			endpointURL := service.URL + endpoint.Path
			relativePath := endpoint.Path

			switch endpoint.HTTPMethod {
			case "GET":
				r.GET(relativePath, handleJSONValidation(endpoint.Body), attachQueryParams(endpoint.QueryParams), proxyRequest(endpointURL))
			case "POST":
				r.POST(relativePath, handleJSONValidation(endpoint.Body), attachQueryParams(endpoint.QueryParams), proxyRequest(endpointURL))
			case "PUT":
				r.PUT(relativePath, handleJSONValidation(endpoint.Body), attachQueryParams(endpoint.QueryParams), proxyRequest(endpointURL))
			case "PATCH":
				r.PATCH(relativePath, handleJSONValidation(endpoint.Body), attachQueryParams(endpoint.QueryParams), proxyRequest(endpointURL))
			}
		}
	}
}

func APIGatewayApp(router internal.RouterInterface, config *internal.Config) (*internal.App, error) {
	middlewares.InitilizeMiddlewares(router, config)
	SetupRoutes(router, config)
	return &internal.App{Router: router}, nil
}
