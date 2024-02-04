package main

import (
	"log"

	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/app"
)

func main() {
	err := internal.InitializeRedis()
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	router := app.InitializeRouter()

	config, err := app.LoadConfig("./config/endpoints.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
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
