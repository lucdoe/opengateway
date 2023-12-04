package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/app"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

func main() {
	err := internal.InitializeRedis()
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	ginRouter := gin.New()
	router := app.GinRouter{Engine: ginRouter}

	middlewares.InitilizeMiddlewares(router.Engine)

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

	APIGatewayApp.Router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
