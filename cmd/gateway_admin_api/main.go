package main

import (
	"log"

	"github.com/gin-gonic/gin"
	i "github.com/lucdoe/gateway_admin_api/internal"
)

func main() {
	i.InitializeRedis()
	db, err := i.InitializePostgres()
	if err != nil {
		log.Fatal(err)
	}
	err = i.MigrateDB(db)
	if err != nil {
		log.Fatal(err)
	}
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on
}
