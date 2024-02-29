package app

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/gateway_admin_api/internal/databases"
	"github.com/lucdoe/gateway_admin_api/internal/service"
	"gorm.io/gorm"
)

type App struct {
	Router *gin.Engine
	DB     *gorm.DB
}

func (app *App) SetupRoutes() *gin.Engine {

	serviceRepo := databases.NewGormServiceRepository(app.DB)
	serviceImpl := service.NewServiceImplementation(serviceRepo)
	serviceController := service.ServiceControllerConstructor(serviceImpl)
	service.SetupServiceRoutes(app.Router, serviceController)

	return app.Router
}

func InitializeAPP(db *gorm.DB) *App {
	app := &App{
		Router: gin.Default(),
		DB:     db,
	}

	app.SetupRoutes()

	return app
}

func (app *App) Run(addr string) {
	log.Fatal(app.Router.Run(addr))
}
