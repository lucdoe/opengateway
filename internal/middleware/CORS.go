package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"
)

type CORS struct {
	AccessControlAllowOrigin      string
	AccessControlAllowCredentials bool
	AccessControlExposeHeaders    string
	AccessControlMaxAge           string
	AccessControlAllowMethods     string
	AccessControlAllowHeaders     string
}

func CORSHandler(c CORS) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", c.AccessControlAllowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", c.AccessControlAllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", c.AccessControlAllowHeaders)
			w.Header().Set("Access-Control-Expose-Headers", c.AccessControlExposeHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", c.AccessControlMaxAge)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (c *CORS) SetAccessControlAllowOrigin(origin string) {
	if origin == "" {
		origin = "*"
		log.Println("CORS: Access-Control-Allow-Origin is not set. Defaulting to '*' (not recommended in production)")
	}
	c.AccessControlAllowOrigin = origin
}

func (c *CORS) SetAccessControlAllowMethods(methods []string) error {
	validMethods := map[string]bool{
		"GET": true, "HEAD": true, "POST": true, "PUT": true,
		"DELETE": true, "CONNECT": true, "OPTIONS": true, "TRACE": true, "PATCH": true,
	}

	var allowedMethods []string
	for _, method := range methods {
		if _, exists := validMethods[method]; exists {
			allowedMethods = append(allowedMethods, method)
		} else {
			return errors.New("Method " + method + " is not a valid HTTP method")
		}
	}

	c.AccessControlAllowMethods = strings.Join(allowedMethods, ",")
	return nil
}

func (c *CORS) SetAccessControllMaxAge(maxAge string) {
	if maxAge == "" {
		log.Println("CORS: Access-Control-Max-Age is not set. Defaulting to 0 (not recommended in production), which disables caching.")
	}
	c.AccessControlMaxAge = maxAge
}

