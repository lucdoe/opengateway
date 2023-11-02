package app

import (
	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
}

func APIGatewayAPP() (*App, error) {
	r := gin.New()

	// loop over yaml to create routes
	return &App{Router: r}, nil
}
