package app

import (
	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

type App struct {
	Router *gin.Engine
}

func APIGatewayAPP() (*App, error) {
	r := gin.New()

	middlewares.InitilizeMiddlewares(r)

	// loop over yaml to create routes
	return &App{Router: r}, nil
}
