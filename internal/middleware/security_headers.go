package middleware

import "net/http"

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

type SecurityHeaders struct {
	ContentSecurityPolicy string
	ContentTypeOptions    string
	FrameOptions          string
	XSSProtection         string
	StrictTransport       string
	ReferrerPolicy        string
	DNSPrefetchControl    string
	FeaturePolicy         string
	PermissionsPolicy     string
}

func SecurityHeadersHandler(s SecurityHeaders) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", s.ContentSecurityPolicy)
			w.Header().Set("X-Content-Type-Options", s.ContentTypeOptions)
			w.Header().Set("X-Frame-Options", s.FrameOptions)
			w.Header().Set("X-XSS-Protection", s.XSSProtection)
			w.Header().Set("Strict-Transport-Security", s.StrictTransport)
			w.Header().Set("Referrer-Policy", s.ReferrerPolicy)
			w.Header().Set("X-DNS-Prefetch-Control", s.DNSPrefetchControl)
			w.Header().Set("Feature-Policy", s.FeaturePolicy)
			w.Header().Set("Permissions-Policy", s.PermissionsPolicy)

			next.ServeHTTP(w, r)
		})
	}
}

func (s *SecurityHeaders) DefaultSecurityHeaders() SecurityHeaders {
	return SecurityHeaders{
		ContentSecurityPolicy: ContentSecurityPolicy,
		ContentTypeOptions:    ContentTypeOptions,
		FrameOptions:          FrameOptions,
		XSSProtection:         XSSProtection,
		StrictTransport:       StrictTransport,
		ReferrerPolicy:        ReferrerPolicy,
		DNSPrefetchControl:    DNSPrefetchControl,
		FeaturePolicy:         FeaturePolicy,
		PermissionsPolicy:     PermissionsPolicy,
	}
}
