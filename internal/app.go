package internal

import (
	"log"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
}

func InitializeAPP() *App {
	app := &App{
		Router: gin.Default(),
	}

	app.SetupRoutes()

	return app
}

func (app *App) InitializeCache() error {
	err := InitializeRedis()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (app *App) InitializeDB() error {
	err := InitializePostgres()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (app *App) Run(addr string) {
	log.Fatal(app.Router.Run(addr))
}
