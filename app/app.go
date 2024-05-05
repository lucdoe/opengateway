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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (app *App) SetupRoutes() *gin.Engine {

	app.Router.Use(CORSMiddleware())

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
