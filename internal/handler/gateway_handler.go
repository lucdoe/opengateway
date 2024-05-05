package handler

import (
	"fmt"
	"net/http"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/middleware"
	"github.com/lucdoe/opengateway/internal/router"
	"github.com/lucdoe/opengateway/internal/utils"
)

func GatewayHandler(cfg *config.TopLevel, mux *http.ServeMux, urlConstructor utils.URLConstructorI, mwHandler middleware.MiddlewareHandlerI, server ServerI) {
	srvcConstructor := router.ServiceRouterConstructor(mux, urlConstructor, mwHandler)
	srvcConstructor.RegisterServices(cfg.Services)

	baseHandler := mwHandler.Add(mux, cfg.GlobalMiddlewares)

	fmt.Println("Starting server on :8080")
	if err := server.ListenAndServe(":8080", baseHandler); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
