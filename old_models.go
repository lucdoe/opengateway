package internal

import (
	"gorm.io/gorm"
)

//lint:file-ignore U1000 Ignore all unused code, it's generated
type Service struct {
	gorm.Model
	name        string
	host        string
	port        string
	enabled     bool
	ProtocolID  uint
	Protocol    Protocol     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Endpoints   []Endpoint   `gorm:"many2many:service_endpoint;"`
	Middlewares []Middleware `gorm:"many2many:service_middleware;"`
	Tags        []Tag        `gorm:"many2many:service_tag;"`
}

type Endpoint struct {
	gorm.Model
	name        string
	enabled     bool
	Paths       []Path       `gorm:"foreignKey:EndpointID;"`
	Headers     []HTTPHeader `gorm:"many2many:endpoint_header;"`
	Middlewares []Middleware `gorm:"many2many:endpoint_middleware;"`
	Tags        []Tag        `gorm:"many2many:endpoint_tag;"`
	Methods     []HTTPMethod `gorm:"many2many:endpoint_method;"`
}

type Middleware struct {
	gorm.Model
	Name        string
	Enabled     bool
	Description string
	Tags        []Tag `gorm:"many2many:middleware_tag;"`
}

type MiddlewareConfiguration struct {
	gorm.Model
	Key          string
	Value        string
	DataType     string
	MiddlewareID uint
	Middleware   Middleware `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Path struct {
	gorm.Model
	name       string
	EndpointID uint
}

type Tag struct {
	gorm.Model
	name string
}

type HTTPMethod struct {
	gorm.Model
	name string
}

type HTTPHeader struct {
	gorm.Model
	name string
}

type Protocol struct {
	gorm.Model
	name     string
	Services []Service `gorm:"foreignKey:ProtocolID"`
}
