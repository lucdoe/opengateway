package app

import (
	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone/internal"
	"github.com/lucdoe/capstone/internal/middlewares"
	monitoring "github.com/zsais/go-gin-prometheus"
)

type App struct {
	Router *gin.Engine
}

func APIGatewayAPP(config *internal.Config) (*App, error) {
	r := gin.New()

	monitoring.NewPrometheus("gin").Use(r)
	middlewares.InitilizeMiddlewares(r)

	for serviceName, service := range config.Services {
		for _, endpoint := range service.Endpoints {
			switch endpoint.HTTPMethod {
			case "GET":
				r.GET(endpoint.Path, middlewares.ValidateRequest(endpoint.AllowedJSON))
			case "POST":
				r.POST(endpoint.Path, middlewares.ValidateRequest(endpoint.AllowedJSON))
			case "PUT":
				r.PUT(endpoint.Path, middlewares.ValidateRequest(endpoint.AllowedJSON))
			case "PATCH":
				r.PATCH(endpoint.Path, middlewares.ValidateRequest(endpoint.AllowedJSON))
			}
			r.Use(middlewares.Proxy(serviceName, endpoint))
		}
	}
	return &App{Router: r}, nil
}
