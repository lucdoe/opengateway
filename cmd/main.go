package main

import (
	"log"

	"github.com/lucdoe/gateway_admin_api/internal"
)

func main() {
	app := internal.InitializeAPP()

	if err := app.InitializeCache(); err != nil {
		log.Fatal("Failed to initialize Cache:", err)
	}

	if err := app.InitializeDB(); err != nil {
		log.Fatal("Failed to initialize Database:", err)
	}

	app.Run(":8080")
}
