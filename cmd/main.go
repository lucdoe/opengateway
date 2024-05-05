package main

import (
	"log"

	"github.com/lucdoe/gateway_admin_api/app"
	"github.com/lucdoe/gateway_admin_api/internal/databases"
)

func main() {
	err := databases.InitializeRedis()
	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}

	db, err := databases.InitializePostgres()
	if err != nil {
		log.Fatal("Failed to initialize Database:", err)
	}

	app := app.InitializeAPP(db)

	app.Run(":8080")
}
