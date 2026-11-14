// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"time"
)

// Settings overrides the global settings for one service, or one endpoint.
// Only the fields it sets count: an endpoint's Settings go over its
// service's, which go over Plugins and Server.
type Settings struct {
	// UpstreamTimeout replaces Server.UpstreamTimeout. It can only be set for
	// a service, since its endpoints share one connection pool.
	UpstreamTimeout time.Duration   `yaml:"UpstreamTimeout"`
	RateLimitConfig RateLimitConfig `yaml:"RateLimit"`
	CORSConfig      CORSConfig      `yaml:"CORS"`
}

// RouteSettings are the plugin settings one endpoint runs with.
type RouteSettings struct {
	RateLimit RateLimitConfig
	CORS      CORSConfig
	// RateLimitScope and CORSScope name the Settings block that changed the
	// plugin's settings for this route: the service, or the service and the
	// endpoint. They are empty when the global settings apply. Routes with
	// the same scope share one instance of the plugin, and one rate limit.
	RateLimitScope string
	CORSScope      string
}

// RouteSettings merges the settings for one endpoint of the named service.
func (c *Config) RouteSettings(serviceName string, endpoint Endpoint) RouteSettings {
	service := c.Services[serviceName]
	endpointScope := serviceName + " " + strings.ToUpper(endpoint.HTTPMethod) + " " + service.Subpath + endpoint.Path

	route := RouteSettings{RateLimit: c.Plugins.RateLimitConfig, CORS: c.Plugins.CORSConfig}
	for _, layer := range []struct {
		settings Settings
		scope    string
	}{{service.Settings, serviceName}, {endpoint.Settings, endpointScope}} {
		if layer.settings.RateLimitConfig != (RateLimitConfig{}) {
			route.RateLimit = layer.settings.RateLimitConfig.over(route.RateLimit)
			route.RateLimitScope = layer.scope
		}
		if layer.settings.CORSConfig != (CORSConfig{}) {
			route.CORS = layer.settings.CORSConfig.over(route.CORS)
			route.CORSScope = layer.scope
		}
	}
	return route
}

// UpstreamTimeout is how long the named service's upstream has to start
// answering.
func (c *Config) UpstreamTimeout(serviceName string) time.Duration {
	if timeout := c.Services[serviceName].Settings.UpstreamTimeout; timeout != 0 {
		return timeout
	}
	return c.Server.UpstreamTimeout
}

// over returns base with the fields r sets replaced.
func (r RateLimitConfig) over(base RateLimitConfig) RateLimitConfig {
	if r.Limit != 0 {
		base.Limit = r.Limit
	}
	if r.Window != 0 {
		base.Window = r.Window
	}
	if r.Key != "" {
		base.Key = r.Key
	}
	return base
}

// over returns base with the fields c sets replaced.
func (c CORSConfig) over(base CORSConfig) CORSConfig {
	if c.Origins != "" {
		base.Origins = c.Origins
	}
	if c.Methods != "" {
		base.Methods = c.Methods
	}
	if c.Headers != "" {
		base.Headers = c.Headers
	}
	return base
}
