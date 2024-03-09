package middleware

import (
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
			w.Header().Set("Access-Control-Allow-Origin", c.getAllowOrigin())
			w.Header().Set("Access-Control-Allow-Methods", getValidMethods(c.AccessControlAllowMethods))
			w.Header().Set("Access-Control-Allow-Headers", c.AccessControlAllowHeaders)
			w.Header().Set("Access-Control-Expose-Headers", c.AccessControlExposeHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", boolToString(c.AccessControlAllowCredentials))
			w.Header().Set("Access-Control-Max-Age", c.AccessControlMaxAge)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (c *CORS) getAllowOrigin() string {
	if c.AccessControlAllowOrigin == "" {
		return "*"
	}
	return c.AccessControlAllowOrigin
}

func getValidMethods(methods string) string {
	validMethods := map[string]bool{
		"GET": true, "HEAD": true, "POST": true, "PUT": true,
		"DELETE": true, "CONNECT": true, "OPTIONS": true, "TRACE": true, "PATCH": true,
	}

	methodsSlice := strings.Split(methods, ",")
	var allowedMethods []string
	for _, method := range methodsSlice {
		if validMethods[strings.ToUpper(method)] {
			allowedMethods = append(allowedMethods, method)
		}
	}

	if len(allowedMethods) == 0 {
		return "GET"
	}

	return strings.Join(allowedMethods, ",")
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
