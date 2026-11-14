// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/plugins/auth"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
	"github.com/lucdoe/opengateway/internal/plugins/cors"
	"github.com/lucdoe/opengateway/internal/plugins/logger"
	ratelimit "github.com/lucdoe/opengateway/internal/plugins/rate-limit"
	"github.com/lucdoe/opengateway/internal/server"
	"github.com/lucdoe/opengateway/internal/store"
)

// storeFactory opens the store shared by the cache and rate-limit plugins.
type storeFactory func(cfg config.StoreConfig) (store.Store, error)

// buildPlugins builds only the plugins some service or endpoint uses, so an
// unused plugin needs no settings and no reachable dependencies.
//
// newStore is called at most once. The returned closers must be closed when
// the gateway stops; if building fails, what was built is already closed.
func buildPlugins(cfg *config.Config, newStore storeFactory) (_ map[string]server.Middleware, closers []io.Closer, err error) {
	plugins := map[string]server.Middleware{}
	defer func() {
		if err != nil {
			for _, closer := range closers {
				closer.Close()
			}
			closers = nil
		}
	}()
	settings := cfg.Plugins

	var shared store.Store
	sharedStore := func() (store.Store, error) {
		if shared != nil {
			return shared, nil
		}
		opened, err := newStore(cfg.Store)
		if err != nil {
			return nil, err
		}
		shared = opened
		return shared, nil
	}

	if cfg.UsesPlugin(config.PluginLogger) {
		requests, err := logger.New(logger.Config{FilePath: settings.LoggerConfig.FilePath})
		if err != nil {
			return nil, closers, fmt.Errorf("%s plugin: %w", config.PluginLogger, err)
		}
		closers = append(closers, requests)
		plugins[config.PluginLogger] = logger.NewMiddleware(requests)
	}

	if cfg.UsesPlugin(config.PluginCORS) {
		policy := cors.New(cors.Config{
			Origins: settings.CORSConfig.Origins,
			Methods: settings.CORSConfig.Methods,
			Headers: settings.CORSConfig.Headers,
		})
		plugins[config.PluginCORS] = cors.NewMiddleware(policy)
	}

	if cfg.UsesPlugin(config.PluginRateLimit) {
		counters, err := sharedStore()
		if err != nil {
			return nil, closers, fmt.Errorf("%s plugin: %w", config.PluginRateLimit, err)
		}
		limiter := ratelimit.NewLimiter(counters, settings.RateLimitConfig.Limit, settings.RateLimitConfig.Window)
		plugins[config.PluginRateLimit] = ratelimit.NewMiddleware(limiter)
	}

	if cfg.UsesPlugin(config.PluginAuth) {
		var keys jwt.Keyfunc
		switch {
		case settings.JWTConfig.JWKSURL != "":
			jwks := auth.NewJWKS(settings.JWTConfig.JWKSURL)
			closers = append(closers, jwks)
			keys = jwks.Keyfunc
		case settings.JWTConfig.SecretKey != "":
			keys = auth.Secret([]byte(settings.JWTConfig.SecretKey))
		default:
			return nil, closers, fmt.Errorf("%s plugin: needs a secret key or a JWKS URL", config.PluginAuth)
		}
		service, err := auth.NewAuthService(auth.JWTConfig{
			Keys:       keys,
			Algorithms: settings.JWTConfig.Algorithms(),
			Issuer:     settings.JWTConfig.Issuer,
			Audience:   settings.JWTConfig.Audience,
			Scope:      settings.JWTConfig.Scope,
		})
		if err != nil {
			return nil, closers, fmt.Errorf("%s plugin: %w", config.PluginAuth, err)
		}
		plugins[config.PluginAuth] = auth.NewMiddleware(service)
	}

	if cfg.UsesPlugin(config.PluginCache) {
		responses, err := sharedStore()
		if err != nil {
			return nil, closers, fmt.Errorf("%s plugin: %w", config.PluginCache, err)
		}
		plugins[config.PluginCache] = cache.NewMiddleware(responses)
	}

	return plugins, closers, nil
}
