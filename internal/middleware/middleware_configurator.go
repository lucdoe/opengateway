package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/logger"
)

type MiddlewareHandlerI interface {
	Add(h http.Handler, middlewares []config.Middleware) http.Handler
}
type MiddlewareHandler struct{}

func MiddlewareHandlerConstructor() MiddlewareHandlerI {
	return &MiddlewareHandler{}
}

type MiddlewareFunc func(http.Handler) http.Handler
type DynamicMiddlewareConfigurator func(cfg *config.Middleware) MiddlewareFunc

var Configurators = map[string]DynamicMiddlewareConfigurator{

	"CORS": func(cfg *config.Middleware) MiddlewareFunc {
		corsConfig := CORS{
			AccessControlAllowOrigin:      cfg.Config.Origins[0],
			AccessControlAllowCredentials: cfg.Config.Credentials,
			AccessControlExposeHeaders:    strings.Join(cfg.Config.Headers, ","),
			AccessControlMaxAge:           fmt.Sprintf("%d", cfg.Config.MaxAge),
			AccessControlAllowMethods:     strings.Join(cfg.Methods, ","),
			AccessControlAllowHeaders:     strings.Join(cfg.Config.Headers, ","),
		}
		return CORSHandler(corsConfig)
	},

	"SecurityHeaders": func(cfg *config.Middleware) MiddlewareFunc {
		securityHeadersConfig := SecurityHeaders{
			XSSProtection:         cfg.Config.XSSProtection,
			ContentTypeOptions:    cfg.Config.ContentTypeOptions,
			FrameOptions:          cfg.Config.FrameOptions,
			ReferrerPolicy:        cfg.Config.ReferrerPolicy,
			FeaturePolicy:         cfg.Config.FeaturePolicy,
			ContentSecurityPolicy: cfg.Config.ContentSecurityPolicy,
			PermissionsPolicy:     cfg.Config.PermissionsPolicy,
		}
		return SecurityHeadersHandler(securityHeadersConfig)
	},

	"BodyLimit": func(cfg *config.Middleware) MiddlewareFunc {
		return BodyLimitHandler(cfg.Config.Max)
	},

	"GZIP": func(cfg *config.Middleware) MiddlewareFunc {
		return func(h http.Handler) http.Handler {
			return GZIPHandler(h)
		}
	},

	"Logger": func(cfg *config.Middleware) MiddlewareFunc {
		return LoggingMiddleware(logger.StandardLogger{})
	},
}

func (m *MiddlewareHandler) Add(h http.Handler, middlewares []config.Middleware) http.Handler {
	for _, mwConfig := range middlewares {

		if mw, ok := Configurators[mwConfig.Name]; ok {
			h = mw(&mwConfig)(h)
		} else {
			fmt.Printf("Middleware '%s' not registered\n", mwConfig.Name)
		}

	}

	return h
}
