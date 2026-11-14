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

package config

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
)

var knownPlugins = []string{PluginLogger, PluginCORS, PluginRateLimit, PluginAuth, PluginCache}

var knownMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete, http.MethodOptions,
}

// supportedSigningMethod is the only JWT signing method the auth plugin
// implements.
const supportedSigningMethod = "HS256"

// Validate checks that the configuration describes a gateway that can start
// and behave as written. It returns one error listing every problem found, so
// they can all be fixed in one pass.
//
// Plugin settings are only checked for plugins that some service or endpoint
// actually uses.
func (c *Config) Validate() error {
	v := &validation{}

	if len(c.Services) == 0 {
		v.addf("no services are configured")
	}

	names := make([]string, 0, len(c.Services))
	for name := range c.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	routes := map[string]string{}
	for _, name := range names {
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
		if slices.Contains(service.Plugins, name) {
			return true
		}
		for _, endpoint := range service.Endpoints {
			if slices.Contains(endpoint.Plugins, name) {
				return true
			}
			if name == PluginAuth && endpoint.Auth.ApplyAuth {
				return true
			}
		}
	}
	return false
}

type validation struct {
	problems []string
}

func (v *validation) addf(format string, args ...any) {
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

// service checks one service and its endpoints. routes maps every route seen
// so far to the endpoint that declared it, to find duplicates.
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

	v.plugins(label, service.Plugins)

	if len(service.Endpoints) == 0 {
		v.addf("%s: at least one endpoint is required", label)
	}

	for i, endpoint := range service.Endpoints {
		endpointLabel := fmt.Sprintf("%s, endpoint #%d", label, i+1)
		if endpoint.Name != "" {
			endpointLabel = fmt.Sprintf("%s, endpoint %q", label, endpoint.Name)
		}

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
		}

		v.plugins(endpointLabel, endpoint.Plugins)

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
		if !slices.Contains(knownPlugins, plugin) {
			v.addf("%s: unknown plugin %q (known plugins: %s)", label, plugin, strings.Join(knownPlugins, ", "))
		}
	}
}

func (v *validation) pluginSettings(c *Config) {
	if c.UsesPlugin(PluginLogger) && c.Plugins.LoggerConfig.FilePath == "" {
		v.addf("Plugins.LoggerConfig.FilePath is required because the logger plugin is used")
	}

	if c.UsesPlugin(PluginCORS) {
		if c.Plugins.CORSConfig.Origins == "" {
			v.addf("Plugins.CORSConfig.Origins is required because the cors plugin is used")
		}
		if c.Plugins.CORSConfig.Methods == "" {
			v.addf("Plugins.CORSConfig.Methods is required because the cors plugin is used")
		}
	}

	if c.UsesPlugin(PluginRateLimit) {
		if c.Plugins.RateLimitConfig.Limit <= 0 {
			v.addf("Plugins.RateLimitConfig.Limit must be greater than 0 because the rate-limit plugin is used")
		}
		if c.Plugins.RateLimitConfig.Window <= 0 {
			v.addf("Plugins.RateLimitConfig.Window must be greater than 0 because the rate-limit plugin is used")
		}
	}

	if c.UsesPlugin(PluginAuth) {
		if c.Plugins.JWTConfig.SecretKey == "" {
			v.addf("Plugins.JWTConfig.SecretKey is required because auth is used")
		}
	}
	if method := c.Plugins.JWTConfig.SigningMethod; method != "" && method != supportedSigningMethod {
		v.addf("Plugins.JWTConfig.SigningMethod %q is not supported; only %s is", method, supportedSigningMethod)
	}
}
