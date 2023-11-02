package main

import (
	"log"

	"github.com/lucdoe/capstone_gateway/internal/app"
	"github.com/lucdoe/capstone_gateway/internal/app/databases"
)

func main() {
	databases.InitializeRedis()

	APIGatewayAPP, err := app.APIGatewayAPP()
	if err != nil {
		log.Fatal(err)
	}

	APIGatewayAPP.Router.Run(":8080")
}
