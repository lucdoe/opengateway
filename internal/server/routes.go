// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
)

// gatewayPath is a route's full path, Subpath and Path, split into what the
// router matches.
type gatewayPath struct {
	// base is the path, or for a prefix path the part before "/*".
	base string
	// prefix reports whether the path ended in "/*", so it also matches
	// every path below base.
	prefix bool
}

func parsePath(full string) gatewayPath {
	if base, ok := strings.CutSuffix(full, config.PrefixSuffix); ok {
		return gatewayPath{base: base, prefix: true}
	}
	return gatewayPath{base: full}
}

// pendingRoute is a route waiting to be registered once all are known, so
// they can be registered most specific first.
type pendingRoute struct {
	path    gatewayPath
	method  string // config.AnyMethod for every method
	handler http.Handler
}

// registerRoutes adds routes to the router so the most specific one wins,
// whatever order the config lists them in: exact paths before prefixes,
// longer prefixes before shorter ones, and a named method before
// config.AnyMethod. The router tries routes in the order they were added.
func (s *Server) registerRoutes(routes []pendingRoute) {
	sort.SliceStable(routes, func(i, j int) bool {
		a, b := routes[i], routes[j]
		if a.path.prefix != b.path.prefix {
			return !a.path.prefix
		}
		if a.path.prefix && len(a.path.base) != len(b.path.base) {
			return len(a.path.base) > len(b.path.base)
		}
		return a.method != config.AnyMethod && b.method == config.AnyMethod
	})

	for _, route := range routes {
		var matched []*mux.Route
		if route.path.base != "" {
			matched = append(matched, s.Router.Handle(route.path.base, route.handler))
		}
		if route.path.prefix {
			matched = append(matched, s.Router.PathPrefix(route.path.base+"/").Handler(route.handler))
		}
		if route.method != config.AnyMethod {
			for _, r := range matched {
				r.Methods(route.method)
			}
		}
	}
}

// pathVariable matches a {name} or {name:pattern} in a path template.
var pathVariable = regexp.MustCompile(`\{([^}:]+)(?::[^}]*)?\}`)

// rewritePath sends next the request with the path upstreamPath makes of
// it: each {name} filled in with the request's value of that path variable,
// and a final "/*" with what came after the prefix of a prefix route.
func rewritePath(path gatewayPath, upstreamPath string, next http.Handler) http.Handler {
	template, rest := strings.CutSuffix(upstreamPath, config.PrefixSuffix)
	baseSegments := strings.Count(path.base, "/")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rewritten := pathVariable.ReplaceAllStringFunc(template, func(variable string) string {
			return vars[pathVariable.FindStringSubmatch(variable)[1]]
		})
		if rest {
			rewritten += remainder(r.URL.Path, baseSegments)
		}

		forwarded := *r
		url := *r.URL
		url.Path, url.RawPath = rewritten, ""
		forwarded.URL = &url
		next.ServeHTTP(w, &forwarded)
	})
}

// remainder is what follows the first segments path segments of path, as
// "/a/b", or "" when nothing does. Counting segments rather than comparing
// text works for prefixes with path variables, such as /users/{id}/*.
func remainder(path string, segments int) string {
	at := 0
	for i := 0; i < segments; i++ {
		next := strings.IndexByte(path[at+1:], '/')
		if next < 0 {
			return ""
		}
		at += next + 1
	}
	return path[at:]
}
