package internal

import (
	"github.com/gin-gonic/gin"
)

func (app *App) SetupRoutes() *gin.Engine {
	r := app.Router

	r = SetupServiceRoutes(r)

	return r
}
