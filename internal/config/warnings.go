// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// exampleSecret is the JWT secret the docs and the example config show.
const exampleSecret = "your-256-bit-secret"

// minSecretBytes is the shortest HS256 secret RFC 7518 allows: as many bits
// as the hash, 256.
const minSecretBytes = 32

// Warnings lists settings that are valid but probably not meant, most of
// them ways tokens can be forged or misused. The gateway still starts; the
// caller prints them.
func (c *Config) Warnings() []string {
	if !c.UsesPlugin(PluginAuth) {
		return nil
	}
	var warnings []string
	jwt := c.Plugins.JWTConfig

	if jwt.SecretKey == exampleSecret {
		warnings = append(warnings, "Plugins.JWTConfig.SecretKey is the example secret from the docs, so anyone can sign tokens with it; set your own, for example with JWT_SECRET")
	} else if len(jwt.SecretKey) < minSecretBytes {
		warnings = append(warnings, fmt.Sprintf("Plugins.JWTConfig.SecretKey is %d bytes; HS256 needs at least %d (256 bits) to resist guessing", len(jwt.SecretKey), minSecretBytes))
	}
	if jwt.Issuer == "" {
		warnings = append(warnings, "Plugins.JWTConfig.Issuer is empty, so tokens from any issuer that has the secret are accepted")
	}
	if jwt.Audience == "" {
		warnings = append(warnings, "Plugins.JWTConfig.Audience is empty, so tokens meant for other services that share the secret are accepted")
	}

	for _, name := range c.ServiceNames() {
		service := c.Services[name]
		if !service.UsesPlugin(PluginAuth) {
			continue
		}
		if upstream, err := url.Parse(service.URL); err == nil && upstream.Scheme == "http" && !internalHost(upstream.Hostname()) {
			warnings = append(warnings, fmt.Sprintf("service %q: tokens are forwarded over plain http to %s; use https", name, upstream.Hostname()))
		}
	}
	return warnings
}

// internalHost reports whether host is reachable only from inside the
// machine or a private network, where plain http is common: loopback and
// private addresses, names without a dot such as a Compose service, and
// names under suffixes reserved for local networks.
func internalHost(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate()
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if !strings.Contains(host, ".") {
		return true
	}
	for _, suffix := range []string{".localhost", ".local", ".internal", ".lan", ".home.arpa", ".svc", ".cluster.local"} {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}
