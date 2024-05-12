package cors

import (
	"strings"
)

type CORSConfig struct {
	Origins string
	Methods string
	Headers string
}

type Cors struct {
	CORSConfig
}

func NewCors(config CORSConfig) *Cors {
	return &Cors{CORSConfig: config}
}

func (c *Cors) ValidateOrigin(origin string) bool {
	if c.Origins == "*" {
		return true
	}
	for _, o := range strings.Split(c.Origins, ", ") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func (c *Cors) ValidateMethod(method string) bool {
	for _, m := range strings.Split(c.Methods, ", ") {
		if strings.TrimSpace(m) == method {
			return true
		}
	}
	return false
}

func (c *Cors) ValidateHeaders(requestedHeaders string) bool {
	if requestedHeaders == "" {
		return true
	}
	allowedHeaders := strings.Split(c.Headers, ", ")
	for _, header := range strings.Split(requestedHeaders, ",") {
		headerFound := false
		for _, allowedHeader := range allowedHeaders {
			if strings.TrimSpace(allowedHeader) == strings.TrimSpace(header) {
				headerFound = true
				break
			}
		}
		if !headerFound {
			return false
		}
	}
	return true
}
