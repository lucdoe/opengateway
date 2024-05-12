package cors

import (
	"net/http"
	"strings"
)

type CORS interface {
	Middleware(next http.Handler) http.Handler
}

type CORSConfig struct {
	Origins string
	Methods string
	Headers string
}

type Cors struct {
	CORSConfig
}

func NewCors(config CORSConfig) CORS {
	return &Cors{
		CORSConfig: config,
	}
}

func (c *Cors) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.validateOrigin(r.Header.Get("Origin")) {
			http.Error(w, "CORS policy violation: Invalid Origin", http.StatusForbidden)
			return
		}

		if r.Method == "OPTIONS" {
			if !c.validateMethod(r.Header.Get("Access-Control-Request-Method")) {
				http.Error(w, "CORS policy violation: Invalid Method", http.StatusForbidden)
				return
			}
			if !c.validateHeaders(r.Header.Get("Access-Control-Request-Headers")) {
				http.Error(w, "CORS policy violation: Invalid Header", http.StatusForbidden)
				return
			}
			c.setCORSHeaders(w, r.Header.Get("Origin"))
			w.WriteHeader(http.StatusOK)
			return
		}

		c.setCORSHeaders(w, r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	})
}

func (c *Cors) validateOrigin(origin string) bool {
	if c.Origins == "*" {
		return true
	}
	allowedOrigins := strings.Split(c.Origins, ", ")
	for _, o := range allowedOrigins {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func (c *Cors) validateMethod(method string) bool {
	allowedMethods := strings.Split(c.Methods, ", ")
	for _, m := range allowedMethods {
		if strings.TrimSpace(m) == method {
			return true
		}
	}
	return false
}

func (c *Cors) validateHeaders(requestedHeaders string) bool {
	if requestedHeaders == "" {
		return true
	}
	headers := strings.Split(requestedHeaders, ",")
	allowedHeaders := strings.Split(c.Headers, ", ")
	for _, header := range headers {
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

func (c *Cors) setCORSHeaders(w http.ResponseWriter, origin string) {
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", c.Methods)
	w.Header().Set("Access-Control-Allow-Headers", c.Headers)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
