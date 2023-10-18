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
			r.Handle(endpoint.HTTPMethod, endpoint.Path, middlewares.Proxy(serviceName, endpoint)).Use(middlewares.ValidateRequest(endpoint.AllowedJSON))
		}
	}
	return &App{Router: r}, nil
}
