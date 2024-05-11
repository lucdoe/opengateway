package cors

import (
	"net/http"
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
		w.Header().Set("Access-Control-Allow-Origin", c.Origins)
		w.Header().Set("Access-Control-Allow-Methods", c.Methods)
		w.Header().Set("Access-Control-Allow-Headers", c.Headers)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		next.ServeHTTP(w, r)
	})
}
