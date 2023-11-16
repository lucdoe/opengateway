package main

import (
	"log"

	"github.com/lucdoe/capstone_gateway/internal/app"
	"github.com/lucdoe/capstone_gateway/internal/app/databases"
)

func main() {
	databases.InitializeRedis()

	fileReader := OSFileReader{}
	yamlParser := YAMLParsing{}
	configLoader := NewConfigLoader(fileReader, yamlParser)

	config, err := configLoader.LoadConfig("gateway_config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	APIGatewayAPP, err := app.APIGatewayAPP(config)
	if err != nil {
		log.Fatal(err)
	}

	APIGatewayAPP.Router.Run(":8080")
}
