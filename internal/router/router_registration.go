package router

import (
	"fmt"
	"net/http"

	"github.com/lucdoe/opengateway/internal/config"
	mw "github.com/lucdoe/opengateway/internal/middleware"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/lucdoe/opengateway/internal/utils"
)

type MuxInterface interface {
	Handle(pattern string, handler http.Handler)
}

type ServiceRouterI interface {
	RegisterServices(services []config.Service)
}

type ServiceRouter struct {
	Mux               MuxInterface
	URLConstructor    utils.URLConstructorI
	MiddlewareHandler mw.MiddlewareHandlerI
}

func ServiceRouterConstructor(mux MuxInterface, urlConstructor utils.URLConstructorI, middlewareHandler mw.MiddlewareHandlerI) *ServiceRouter {
	return &ServiceRouter{
		Mux:               mux,
		URLConstructor:    urlConstructor,
		MiddlewareHandler: middlewareHandler,
	}
}

func (sr *ServiceRouter) RegisterServices(services []config.Service) {
	for _, service := range services {
		sr.registerEndpoints(service)
	}
}

func (sr *ServiceRouter) registerEndpoints(service config.Service) {
	serviceBaseURL := fmt.Sprintf("%s://%s:%d", service.Protocol, service.Host, service.Port)

	for _, endpoint := range service.Endpoints {
		fullPath, err := sr.URLConstructor.ConstructURL(serviceBaseURL, service.BasePath+endpoint.Path)
		if err != nil {
			fmt.Printf("Error parsing service URL '%s': %s\n", serviceBaseURL+service.BasePath+endpoint.Path, err)
			continue
		}

		reverseProxyHandler := proxy.ReverseProxyHandler(fullPath)
		wrappedHandler := sr.MiddlewareHandler.Add(reverseProxyHandler, endpoint.Middlewares)

		sr.Mux.Handle(service.BasePath+endpoint.Path, wrappedHandler)
	}
}
