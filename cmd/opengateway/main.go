package main

import (
	"fmt"
	"net/http"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/handler"
	"github.com/lucdoe/opengateway/internal/middleware"
	"github.com/lucdoe/opengateway/internal/utils"
)

func main() {
	fileReader := &config.OSFileReader{}
	loader := config.NewLoader(fileReader)

	cfg, err := loader.LoadConfigFromFile("config.yaml")
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	mux := http.NewServeMux()
	urlConstructor := utils.GatewayURLConstructor()
	mwHandler := middleware.MiddlewareHandlerConstructor()
	server := handler.APIGatewayServer()

	handler.GatewayHandler(cfg, mux, urlConstructor, mwHandler, server)
}
