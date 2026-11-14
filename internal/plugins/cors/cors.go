// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// NewMiddleware returns the cors plugin: it rejects requests from origins the
// policy does not allow, answers preflight requests and sets the CORS
// response headers.
func NewMiddleware(p *Policy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if !p.AllowsOrigin(origin) {
				http.Error(w, "CORS policy violation: Invalid Origin", http.StatusForbidden)
				return
			}

			preflight := r.Method == http.MethodOptions
			if preflight && (!p.AllowsMethod(r.Header.Get("Access-Control-Request-Method")) ||
				!p.AllowsHeaders(r.Header.Get("Access-Control-Request-Headers"))) {
				http.Error(w, "CORS policy violation: Invalid Method or Header", http.StatusForbidden)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", p.allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", p.allowHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
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
