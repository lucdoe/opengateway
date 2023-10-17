package app

import (
	"github.com/gin-gonic/gin"
	monitoring "github.com/zsais/go-gin-prometheus"

	"github.com/lucdoe/capstone/internal/middlewares"
)

type App struct {
	Router *gin.Engine
}

func APIGatewayAPP() (*App, error) {
	r := gin.New()

	monitoring.NewPrometheus("gin").Use(r)
	middlewares.InitilizeMiddlewares(r)

	d := r.Group("/definitions")
	DefinitionRoutes(d)

	s := r.Group("/schemas")
	SchemaRoutes(s)

	a := r.Group("/authors")
	AuthorRoutes(a)

	p := r.Group("/partners")
	PartnerRoutes(p)

	c := r.Group("/cases")
	CaseRoutes(c)

	u := r.Group("/users")
	UserRoutes(u)

	return &App{Router: r}, nil
}
