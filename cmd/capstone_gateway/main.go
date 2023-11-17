package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal/app"
	"github.com/lucdoe/capstone_gateway/internal/app/router"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
	"github.com/lucdoe/capstone_gateway/internal/utils"
)

func main() {
	ginRouter := gin.New()
	router := router.GinRouter{Engine: ginRouter}

	middlewares.InitilizeMiddlewares(ginRouter)

	fileReader := utils.OSFileReader{}
	yamlParser := utils.YAMLParsing{}
	configLoader := utils.NewConfigLoader(fileReader, yamlParser)

	config, err := configLoader.LoadConfig("endpoints.yaml")
	if err != nil {
		log.Fatal(err)
	}

	APIGatewayApp, err := app.APIGatewayApp(router, config)
	if err != nil {
		log.Fatal(err)
	}

	if err := APIGatewayApp.Router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
