package internal

import "gorm.io/gorm"

//lint:file-ignore U1000 Ignore all unused code, it's generated
type Service struct {
	gorm.Model
	ServiceID   string `gorm:"primaryKey"`
	name        string
	host        string
	port        string
	enabled     bool
	Protocol    Protocol
	Endpoints   []Endpoint   `gorm:"many2many:service_endpoint;"`
	Middlewares []Middleware `gorm:"many2many:service_middleware;"`
	Tags        []Tag        `gorm:"many2many:service_tag;"`
}

type Endpoint struct {
	gorm.Model
	EndpointID  string `gorm:"primaryKey"`
	enabled     bool
	Paths       []Path
	Headers     []HTTPHeader `gorm:"many2many:endpoint_header;"`
	Middlewares []Middleware `gorm:"many2many:endpoint_middleware;"`
	Tags        []Tag        `gorm:"many2many:endpoint_tag;"`
}

type Middleware struct {
	gorm.Model
	MiddlewareID string `gorm:"primaryKey"`
	name         string
	enabled      bool
	description  string
	config       MiddlewareConfiguration
	Tags         []Tag `gorm:"many2many:middleware_tag;"`
}

type MiddlewareConfiguration struct {
	gorm.Model
	ConfigID string `gorm:"primaryKey"`
	key      string
	value    string
	DataType string
}

type Path struct {
	gorm.Model
	PathID string `gorm:"primaryKey"`
	name   string
}

type Tag struct {
	gorm.Model
	TagID string `gorm:"primaryKey"`
	name  string
}

type HTTPMethod struct {
	// ID is method name
	MethodID string `gorm:"primaryKey"`
}

type HTTPHeader struct {
	// ID is header name
	HeaderID string `gorm:"primaryKey"`
}

type Protocol struct {
	ProtocolID string `gorm:"primaryKey"`
}
