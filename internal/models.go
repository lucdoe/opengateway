package internal

import (
	"gorm.io/gorm"
)

type Service struct {
	gorm.Model
	Name        string
	Host        string
	Port        string
	Enabled     bool
	ProtocolID  uint
	Protocol    Protocol     `gorm:"foreignKey:ProtocolID"`
	Endpoints   []Endpoint   `gorm:"foreignKey:ServiceID"`
	Middlewares []Middleware `gorm:"many2many:service_middlewares;foreignKey:ID;joinForeignKey:ServiceID;References:ID;JoinReferences:MiddlewareID"`
	Tags        []Tag        `gorm:"many2many:service_tags;foreignKey:ID;joinForeignKey:ServiceID;References:ID;JoinReferences:TagID"`
}

type Endpoint struct {
	gorm.Model
	ServiceID   uint
	Name        string
	Enabled     bool
	Paths       []Path       `gorm:"foreignKey:EndpointID"`
	Headers     []HTTPHeader `gorm:"many2many:endpoint_http_headers;foreignKey:ID;joinForeignKey:EndpointID;References:ID;JoinReferences:HTTPHeaderID"`
	Middlewares []Middleware `gorm:"many2many:endpoint_middlewares;foreignKey:ID;joinForeignKey:EndpointID;References:ID;JoinReferences:MiddlewareID"`
	Tags        []Tag        `gorm:"many2many:endpoint_tags;foreignKey:ID;joinForeignKey:EndpointID;References:ID;JoinReferences:TagID"`
}

type Path struct {
	gorm.Model
	EndpointID  uint
	Name        string
	HTTPMethods []HTTPMethod `gorm:"many2many:path_http_methods;foreignKey:ID;joinForeignKey:PathID;References:ID;JoinReferences:HTTPMethodID"`
}

type Middleware struct {
	gorm.Model
	Name        string
	Enabled     bool
	Description string
	Config      MiddlewareConfig `gorm:"foreignKey:MiddlewareID"`
	Tags        []Tag            `gorm:"many2many:middleware_tags;foreignKey:ID;joinForeignKey:MiddlewareID;References:ID;JoinReferences:TagID"`
}

type MiddlewareConfig struct {
	gorm.Model
	Key          string
	Value        string
	DataType     string
	MiddlewareID uint
}

type HTTPHeader struct {
	gorm.Model
	Name string
}

type HTTPMethod struct {
	gorm.Model
	Name string
}

type Tag struct {
	gorm.Model
	Name string
}

type Protocol struct {
	gorm.Model
	Name     string
	Services []Service `gorm:"foreignKey:ProtocolID"`
}
