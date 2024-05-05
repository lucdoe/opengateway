package middlewares

import (
	"github.com/gin-gonic/gin"
)

const (
	ContentSecurityPolicy = "default-src 'self'; frame-ancestors 'none'; script-src 'self'; object-src 'none'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;"
	ContentTypeOptions    = "nosniff"
	FrameOptions          = "DENY"
	XSSProtection         = "1; mode=block"
	StrictTransport       = "max-age=63072000; includeSubDomains"
	ReferrerPolicy        = "strict-origin-when-cross-origin"
	DNSPrefetchControl    = "off"
	FeaturePolicy         = "geolocation 'none'; midi 'none'; notifications 'none'; push 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; speaker 'none'; vibrate 'none'; fullscreen 'self'; payment 'none';"
	PermissionsPolicy     = "geolocation=(), midi=(), notifications=(), push=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), payment=()"
)

func SecurityHeaders(c *gin.Context) {
	c.Header("Content-Security-Policy", ContentSecurityPolicy)
	c.Header("X-Content-Type-Options", ContentTypeOptions)
	c.Header("X-Frame-Options", FrameOptions)
	c.Header("X-XSS-Protection", XSSProtection)
	c.Header("Strict-Transport-Security", StrictTransport)
	c.Header("Referrer-Policy", ReferrerPolicy)
	c.Header("X-DNS-Prefetch-Control", DNSPrefetchControl)
	c.Header("Feature-Policy", FeaturePolicy)
	c.Header("Permissions-Policy", PermissionsPolicy)

	c.Next()
}
