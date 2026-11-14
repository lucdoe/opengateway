// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

var knownMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete, http.MethodOptions,
}

// ReservedPathPrefix starts the paths the gateway answers itself, such as
// /__health, so no service can route them.
const ReservedPathPrefix = "/__"

// Validate reports every problem in one error, so they can be fixed in one
// pass. Plugin settings are only checked for plugins in use.
func (c *Config) Validate() error {
	v := &validation{}

	if !validPort(c.Server.Port) {
		v.addf("Server.Port %d must be between 1 and 65535", c.Server.Port)
	}
	for _, timeout := range []struct {
		name  string
		value time.Duration
	}{
		{"ReadHeaderTimeout", c.Server.ReadHeaderTimeout},
		{"IdleTimeout", c.Server.IdleTimeout},
		{"ShutdownTimeout", c.Server.ShutdownTimeout},
		{"UpstreamTimeout", c.Server.UpstreamTimeout},
		{"CircuitBreaker.OpenTimeout", c.Server.CircuitBreaker.OpenTimeout},
	} {
		if timeout.value <= 0 {
			v.addf("Server.%s %s must be greater than 0", timeout.name, timeout.value)
		}
	}
	for _, count := range []struct {
		name  string
		value int
	}{
		{"ConsecutiveFailures", c.Server.CircuitBreaker.ConsecutiveFailures},
		{"HalfOpenRequests", c.Server.CircuitBreaker.HalfOpenRequests},
	} {
		if count.value <= 0 {
			v.addf("Server.CircuitBreaker.%s %d must be greater than 0", count.name, count.value)
		}
	}
	switch c.Store.Type {
	case StoreMemory:
	case StoreRedis:
		if c.UsesPlugin(PluginCache) || c.UsesPlugin(PluginRateLimit) {
			if c.Store.Redis.Host == "" {
				v.addf("Store.Redis.Host is required because Store.Type is redis")
			}
			if !validPort(c.Store.Redis.Port) {
				v.addf("Store.Redis.Port %d must be between 1 and 65535", c.Store.Redis.Port)
			}
		}
	default:
		v.addf("Store.Type %q is not one of %s, %s", c.Store.Type, StoreMemory, StoreRedis)
	}

	if len(c.Services) == 0 {
		v.addf("no services are configured")
	}

	routes := map[string]string{}
	for _, name := range c.ServiceNames() {
		v.service(name, c.Services[name], routes)
	}

	v.pluginSettings(c)

	if len(v.problems) == 0 {
		return nil
	}
	return fmt.Errorf("invalid configuration:\n- %s", strings.Join(v.problems, "\n- "))
}

// UsesPlugin reports whether any service or endpoint enables the named
// plugin. Auth also counts as used when an endpoint sets ApplyAuth.
func (c *Config) UsesPlugin(name string) bool {
	for _, service := range c.Services {
		if service.UsesPlugin(name) {
			return true
		}
	}
	return false
}

// UsesPlugin reports whether the service or one of its endpoints enables the
// named plugin, as Config.UsesPlugin does.
func (s Service) UsesPlugin(name string) bool {
	if slices.Contains(s.Plugins, name) {
		return true
	}
	for _, endpoint := range s.Endpoints {
		if slices.Contains(endpoint.Plugins, name) || (name == PluginAuth && endpoint.Auth.ApplyAuth) {
			return true
		}
	}
	return false
}

// headerName matches the token a header name is made of (RFC 9110).
var headerName = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")

func validRateLimitKey(key string) bool {
	if name, isHeader := strings.CutPrefix(key, RateLimitByHeader); isHeader {
		return headerName.MatchString(name)
	}
	return key == "" || key == RateLimitByIP || key == RateLimitBySubject
}

func validPort(port int) bool {
	return port >= 1 && port <= 65535
}

type validation struct {
	problems []string
}

func (v *validation) addf(format string, args ...any) {
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

// service checks one service and its endpoints. routes maps every route seen
// so far to the endpoint that declared it, to find duplicates.
// endpointLabel names the i-th endpoint of a service in error messages, by
// its Name if it has one.
func endpointLabel(service string, i int, endpoint Endpoint) string {
	if endpoint.Name != "" {
		return fmt.Sprintf("service %q, endpoint %q", service, endpoint.Name)
	}
	return fmt.Sprintf("service %q, endpoint #%d", service, i+1)
}

func (v *validation) service(name string, service Service, routes map[string]string) {
	label := fmt.Sprintf("service %q", name)

	if upstream, err := url.Parse(service.URL); service.URL == "" {
		v.addf("%s: URL is required", label)
	} else if err != nil || (upstream.Scheme != "http" && upstream.Scheme != "https") || upstream.Host == "" {
		v.addf("%s: URL %q must be an absolute http or https URL", label, service.URL)
	}

	if service.Subpath != "" && !strings.HasPrefix(service.Subpath, "/") {
		v.addf("%s: Subpath %q must start with \"/\"", label, service.Subpath)
	}
	// Paths start with "/", so a trailing one would make every route "//".
	if strings.HasSuffix(service.Subpath, "/") {
		v.addf("%s: Subpath %q must not end with \"/\"", label, service.Subpath)
	}

	v.plugins(label, service.Plugins)

	if len(service.Endpoints) == 0 {
		v.addf("%s: at least one endpoint is required", label)
	}

	for i, endpoint := range service.Endpoints {
		endpointLabel := endpointLabel(name, i, endpoint)

		method := strings.ToUpper(endpoint.HTTPMethod)
		validMethod := slices.Contains(knownMethods, method)
		if endpoint.HTTPMethod == "" {
			v.addf("%s: HTTPMethod is required", endpointLabel)
		} else if !validMethod {
			v.addf("%s: HTTPMethod %q is not one of %s", endpointLabel, endpoint.HTTPMethod, strings.Join(knownMethods, ", "))
		}

		validPath := strings.HasPrefix(endpoint.Path, "/")
		if endpoint.Path == "" {
			v.addf("%s: Path is required", endpointLabel)
		} else if !validPath {
			v.addf("%s: Path %q must start with \"/\"", endpointLabel, endpoint.Path)
		} else if full := service.Subpath + endpoint.Path; strings.HasPrefix(full, ReservedPathPrefix) {
			v.addf("%s: path %s is reserved: paths starting with %s belong to the gateway", endpointLabel, full, ReservedPathPrefix)
		}

		v.plugins(endpointLabel, endpoint.Plugins)

		usesAuth := endpoint.Auth.ApplyAuth || slices.Contains(service.Plugins, PluginAuth) || slices.Contains(endpoint.Plugins, PluginAuth)
		if endpoint.Auth.Scope != "" && !usesAuth {
			v.addf("%s: Auth.Scope needs auth: set Auth.ApplyAuth or add the auth plugin", endpointLabel)
		}

		if validMethod && validPath {
			route := method + " " + service.Subpath + endpoint.Path
			if first, taken := routes[route]; taken {
				v.addf("%s: route %s is already used by %s", endpointLabel, route, first)
			} else {
				routes[route] = endpointLabel
			}
		}
	}
}

func (v *validation) plugins(label string, plugins []string) {
	for _, plugin := range plugins {
		if !slices.Contains(PluginOrder, plugin) {
			v.addf("%s: unknown plugin %q (known plugins: %s)", label, plugin, strings.Join(PluginOrder, ", "))
		}
	}
}

// pluginSettings checks the settings every route runs with, after its
// service's and its own Settings are merged in. A problem in the global
// settings is reported once, under Plugins; one a Settings block caused,
// under the endpoint.
func (v *validation) pluginSettings(c *Config) {
	reported := map[string]bool{}
	once := func(format string, args ...any) {
		if message := fmt.Sprintf(format, args...); !reported[message] {
			reported[message] = true
			v.addf("%s", message)
		}
	}

	for _, name := range c.ServiceNames() {
		service := c.Services[name]
		if service.Settings.UpstreamTimeout < 0 {
			v.addf("service %q: Settings.UpstreamTimeout %s must be greater than 0", name, service.Settings.UpstreamTimeout)
		}
		for i, endpoint := range service.Endpoints {
			label := endpointLabel(name, i, endpoint)
			if endpoint.Settings.UpstreamTimeout != 0 {
				v.addf("%s: Settings.UpstreamTimeout can only be set for the service, whose endpoints share one connection pool", label)
			}
			route := c.RouteSettings(name, endpoint)
			uses := func(plugin string) bool {
				return slices.Contains(service.Plugins, plugin) || slices.Contains(endpoint.Plugins, plugin)
			}

			if uses(PluginCORS) {
				where := "Plugins.CORSConfig"
				if route.CORSScope != "" {
					where = label + ": CORSConfig"
				}
				if route.CORS.Origins == "" {
					once("%s.Origins is required because the cors plugin is used", where)
				}
				if route.CORS.Methods == "" {
					once("%s.Methods is required because the cors plugin is used", where)
				}
			}

			if uses(PluginRateLimit) {
				where := "Plugins.RateLimitConfig"
				if route.RateLimitScope != "" {
					where = label + ": RateLimitConfig"
				}
				if route.RateLimit.Limit <= 0 {
					once("%s.Limit must be greater than 0 because the rate-limit plugin is used", where)
				}
				if route.RateLimit.Window <= 0 {
					once("%s.Window must be greater than 0 because the rate-limit plugin is used", where)
				}
				if key := route.RateLimit.Key; !validRateLimitKey(key) {
					once("%s.Key %q is not one of %s, %sName-Of-Header or %s", where, key, RateLimitByIP, RateLimitByHeader, RateLimitBySubject)
				}
			}
		}
	}

	if c.UsesPlugin(PluginAuth) {
		v.jwt(c.Plugins.JWTConfig)
	}
}

func (v *validation) jwt(jwt JWTConfig) {
	switch {
	case jwt.SecretKey == "" && jwt.JWKSURL == "":
		v.addf("Plugins.JWTConfig needs SecretKey (for HS256 tokens) or JWKSURL (for RS256, ES256 and other tokens from an identity provider) because auth is used")
	case jwt.SecretKey != "" && jwt.JWKSURL != "":
		v.addf("Plugins.JWTConfig sets both SecretKey and JWKSURL; set one (JWT_SECRET also sets SecretKey)")
	case jwt.JWKSURL != "":
		if keys, err := url.Parse(jwt.JWKSURL); err != nil || (keys.Scheme != "https" && keys.Scheme != "http") || keys.Host == "" {
			v.addf("Plugins.JWTConfig.JWKSURL %q must be an absolute https URL", jwt.JWKSURL)
		} else if keys.Scheme == "http" && !internalHost(keys.Hostname()) {
			v.addf("Plugins.JWTConfig.JWKSURL %q must use https: keys fetched over plain http can be swapped to forge tokens", jwt.JWKSURL)
		}
	}

	if method := jwt.SigningMethod; method != "" {
		allowed := HMACAlgorithms
		if jwt.JWKSURL != "" {
			allowed = KeyAlgorithms
		}
		if !slices.Contains(allowed, method) {
			v.addf("Plugins.JWTConfig.SigningMethod %q is not one of %s", method, strings.Join(allowed, ", "))
		}
	}
}
