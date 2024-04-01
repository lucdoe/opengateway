package server

import "net/http"

type HandlerFunc func()

type Router interface {
	Use(...HandlerFunc) Router

	Any(path string, handlers ...HandlerFunc) Router
	GET(path string, handlers ...HandlerFunc) Router
	POST(path string, handlers ...HandlerFunc) Router
	DELETE(path string, handlers ...HandlerFunc) Router
	PATCH(path string, handlers ...HandlerFunc) Router
	PUT(path string, handlers ...HandlerFunc) Router
	OPTIONS(path string, handlers ...HandlerFunc) Router
	HEAD(path string, handlers ...HandlerFunc) Router

	Run(addr ...string)
}

type concreteGinRouter struct {
	router GinRouter
}

type GinRouter interface {
	Use(...HandlerFunc) Router

	Any(method string, path string, handlers ...HandlerFunc) Router
	GET(method string, path string, handlers ...HandlerFunc) Router
	POST(method string, path string, handlers ...HandlerFunc) Router
	DELETE(method string, path string, handlers ...HandlerFunc) Router
	PATCH(method string, path string, handlers ...HandlerFunc) Router
	PUT(method string, path string, handlers ...HandlerFunc) Router
	OPTIONS(method string, path string, handlers ...HandlerFunc) Router
	HEAD(method string, path string, handlers ...HandlerFunc) Router

	Run(addr ...string)
}

func NewGinRouter(router GinRouter) Router {
	return &concreteGinRouter{router: router}
}

func (r *concreteGinRouter) Use(handlers ...HandlerFunc) Router {
	return r.router.Use()
}

func (r *concreteGinRouter) Any(path string, handlers ...HandlerFunc) Router {
	return r.router.Any(http.MethodGet, path, handlers...)
}

func (r *concreteGinRouter) GET(path string, handlers ...HandlerFunc) Router {
	return r.router.GET(http.MethodGet, path, handlers...)
}

func (r *concreteGinRouter) POST(path string, handlers ...HandlerFunc) Router {
	return r.router.POST(http.MethodPost, path, handlers...)
}

func (r *concreteGinRouter) DELETE(path string, handlers ...HandlerFunc) Router {
	return r.router.DELETE(http.MethodDelete, path, handlers...)
}

func (r *concreteGinRouter) PATCH(path string, handlers ...HandlerFunc) Router {
	return r.router.PATCH(http.MethodPatch, path, handlers...)
}

func (r *concreteGinRouter) PUT(path string, handlers ...HandlerFunc) Router {
	return r.router.PUT(http.MethodPut, path, handlers...)
}

func (r *concreteGinRouter) OPTIONS(path string, handlers ...HandlerFunc) Router {
	return r.router.OPTIONS(http.MethodOptions, path, handlers...)
}

func (r *concreteGinRouter) HEAD(path string, handlers ...HandlerFunc) Router {
	return r.router.HEAD(http.MethodHead, path, handlers...)
}

func (r *concreteGinRouter) Run(addr ...string) {
	r.router.Run()
}
