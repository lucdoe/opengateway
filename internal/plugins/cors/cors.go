// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package cors

import (
	"net/http"
	"slices"
	"strings"
)

// Config lists what the policy allows, comma-separated. An Origins entry of
// "*" allows every origin.
type Config struct {
	Origins string
	Methods string
	Headers string
}

// Policy is a parsed Config.
type Policy struct {
	anyOrigin    bool
	origins      []string
	methods      []string
	headers      []string // lower case: header names are case-insensitive
	allowMethods string
	allowHeaders string
}

func New(cfg Config) *Policy {
	origins := split(cfg.Origins)
	return &Policy{
		anyOrigin:    slices.Contains(origins, "*"),
		origins:      origins,
		methods:      split(cfg.Methods),
		headers:      split(strings.ToLower(cfg.Headers)),
		allowMethods: cfg.Methods,
		allowHeaders: cfg.Headers,
	}
}

func (p *Policy) AllowsOrigin(origin string) bool {
	return p.anyOrigin || slices.Contains(p.origins, origin)
}

func (p *Policy) AllowsMethod(method string) bool {
	return slices.Contains(p.methods, method)
}

// AllowsHeaders reports whether every header in a comma-separated
// Access-Control-Request-Headers value is allowed.
func (p *Policy) AllowsHeaders(requested string) bool {
	for _, header := range split(strings.ToLower(requested)) {
		if !slices.Contains(p.headers, header) {
			return false
		}
	}
	return true
}

// NewMiddleware returns the cors plugin. A request without an Origin header
// is not a CORS request and passes untouched. Otherwise it rejects origins
// the policy does not allow, answers preflight requests and sets the CORS
// response headers.
//
// With "*" every origin gets Access-Control-Allow-Origin: * and no
// credentials, so other sites cannot make requests with the user's cookies.
// A listed origin is echoed back with credentials allowed, and every
// response carries Vary: Origin, since its CORS headers depend on it.
func NewMiddleware(p *Policy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !p.anyOrigin {
				w.Header().Add("Vary", "Origin")
			}
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			if !p.AllowsOrigin(origin) {
				http.Error(w, "CORS policy violation: Invalid Origin", http.StatusForbidden)
				return
			}

			preflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
			if preflight && (!p.AllowsMethod(r.Header.Get("Access-Control-Request-Method")) ||
				!p.AllowsHeaders(r.Header.Get("Access-Control-Request-Headers"))) {
				http.Error(w, "CORS policy violation: Invalid Method or Header", http.StatusForbidden)
				return
			}

			if p.anyOrigin {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Methods", p.allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", p.allowHeaders)
			if preflight {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func split(list string) []string {
	var items []string
	for _, item := range strings.Split(list, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
