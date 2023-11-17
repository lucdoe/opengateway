package app

import "github.com/gin-gonic/gin"

type MockRouter struct {
	registeredRoutes   map[string][]gin.HandlerFunc
	appliedMiddlewares []gin.HandlerFunc
}

func NewMockRouter() *MockRouter {
	return &MockRouter{
		registeredRoutes:   make(map[string][]gin.HandlerFunc),
		appliedMiddlewares: []gin.HandlerFunc{},
	}
}

func (m *MockRouter) GET(path string, handlers ...gin.HandlerFunc) {
	m.registeredRoutes["GET "+path] = handlers
}

func (m *MockRouter) POST(path string, handlers ...gin.HandlerFunc) {
	m.registeredRoutes["POST "+path] = handlers
}

func (m *MockRouter) PATCH(path string, handlers ...gin.HandlerFunc) {
	m.registeredRoutes["PATCH "+path] = handlers
}

func (m *MockRouter) PUT(path string, handlers ...gin.HandlerFunc) {
	m.registeredRoutes["PUT "+path] = handlers
}

func (m *MockRouter) Run(addr ...string) error {
	// No-op for testing
	return nil
}

func (m *MockRouter) Use(middlewares ...gin.HandlerFunc) {
	m.appliedMiddlewares = append(m.appliedMiddlewares, middlewares...)
}
