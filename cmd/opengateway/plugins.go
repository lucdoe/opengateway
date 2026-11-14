// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// routePlugins are the plugins the gateway runs: one instance of each, and
// for routes whose Settings change cors or rate-limit, an instance of that
// plugin per Settings block, built when the route is set up.
type routePlugins struct {
	defaults server.Middlewares
	counters store.Store // for rate-limit instances of a Settings block
	scoped   map[string]server.Middleware
}

func (p *routePlugins) Middleware(name string, route config.RouteSettings) (server.Middleware, bool) {
	mw, ok := p.defaults[name]
	scope := map[string]string{config.PluginRateLimit: route.RateLimitScope, config.PluginCORS: route.CORSScope}[name]
	if !ok || scope == "" {
		return mw, ok
	}
	id := name + " " + scope
	if built, ok := p.scoped[id]; ok {
		return built, true
	}
	if name == config.PluginRateLimit {
		mw = rateLimitPlugin(p.counters, route.RateLimit, scope)
	} else {
		mw = corsPlugin(route.CORS)
	}
	p.scoped[id] = mw
	return mw, true
}

func corsPlugin(settings config.CORSConfig) server.Middleware {
	return cors.NewMiddleware(cors.New(cors.Config{Origins: settings.Origins, Methods: settings.Methods, Headers: settings.Headers}))
}

// rateLimitPlugin counts with settings. Counters of a scope other than ""
// get their own keys, so a Settings block's limit and the global one never
// share a count; the global keys stay as they were.
func rateLimitPlugin(counters store.Store, settings config.RateLimitConfig, scope string) server.Middleware {
	key := rateLimitKey(settings.Key)
	if scope != "" {
		sum := sha256.Sum256([]byte(scope))
		prefix, unscoped := hex.EncodeToString(sum[:8])+":", key
		key = func(r *http.Request) string { return prefix + unscoped(r) }
	}
	return ratelimit.NewMiddleware(ratelimit.NewLimiter(counters, settings.Limit, settings.Window), key)
}

// buildPlugins builds only the plugins some service or endpoint uses, so an
// unused plugin needs no settings and no reachable dependencies.
//
// newStore is called at most once. The returned closers must be closed when
// the gateway stops; if building fails, what was built is already closed.
func buildPlugins(cfg *config.Config, newStore storeFactory) (_ *routePlugins, closers []io.Closer, err error) {
	plugins := &routePlugins{defaults: server.Middlewares{}, scoped: map[string]server.Middleware{}}
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
		plugins.defaults[config.PluginLogger] = logger.NewMiddleware(requests)
	}

	if cfg.UsesPlugin(config.PluginCORS) {
		plugins.defaults[config.PluginCORS] = corsPlugin(settings.CORSConfig)
	}

	if cfg.UsesPlugin(config.PluginRateLimit) {
		counters, err := sharedStore()
		if err != nil {
			return nil, closers, fmt.Errorf("%s plugin: %w", config.PluginRateLimit, err)
		}
		plugins.counters = counters
		plugins.defaults[config.PluginRateLimit] = rateLimitPlugin(counters, settings.RateLimitConfig, "")
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
		plugins.defaults[config.PluginAuth] = auth.NewMiddleware(service)
	}

	if cfg.UsesPlugin(config.PluginCache) {
		responses, err := sharedStore()
		if err != nil {
			return nil, closers, fmt.Errorf("%s plugin: %w", config.PluginCache, err)
		}
		plugins.defaults[config.PluginCache] = cache.NewMiddleware(responses)
	}

	return plugins, closers, nil
}

// rateLimitKey returns what the rate-limit plugin counts by, for a
// validated RateLimitConfig.Key.
func rateLimitKey(key string) ratelimit.KeyFunc {
	if header, ok := strings.CutPrefix(key, config.RateLimitByHeader); ok {
		return ratelimit.ByHeader(header)
	}
	if key == config.RateLimitBySubject {
		// Rate-limit runs after auth with this key (see
		// config.PluginOrderFor), so the claims are verified ones.
		return ratelimit.ByValue("sub", func(r *http.Request) string {
			if claims := auth.ClaimsFrom(r.Context()); claims != nil {
				return claims.Subject
			}
			return ""
		})
	}
	return ratelimit.ByClientIP
}
