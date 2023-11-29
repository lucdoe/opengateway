package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/app"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

func main() {
	ginRouter := gin.New()
	router := app.GinRouter{Engine: ginRouter}

	middlewares.InitilizeMiddlewares(ginRouter)

	fileReader := internal.OSFileReader{}
	yamlParser := internal.YAMLParsing{}
	configLoader := internal.NewConfigLoader(fileReader, yamlParser)

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
