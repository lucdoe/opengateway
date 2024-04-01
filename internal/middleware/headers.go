package middleware

import "net/http"

type SecurityMiddleware interface {
	SetHeaders(next http.Handler) http.Handler
}

type SecurityHeadersPolicy struct {
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

type concreteSecurityMiddleware struct {
	policy SecurityHeadersPolicy
}

func NewSecurityHeadersMiddleware(policy SecurityHeadersPolicy) SecurityMiddleware {
	return &concreteSecurityMiddleware{
		policy: policy,
	}
}

func (sm *concreteSecurityMiddleware) SetHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", sm.policy.ContentSecurityPolicy)
		w.Header().Set("X-Content-Type-Options", sm.policy.ContentTypeOptions)
		w.Header().Set("X-Frame-Options", sm.policy.FrameOptions)
		w.Header().Set("X-XSS-Protection", sm.policy.XSSProtection)
		w.Header().Set("Strict-Transport-Security", sm.policy.StrictTransport)
		w.Header().Set("Referrer-Policy", sm.policy.ReferrerPolicy)
		w.Header().Set("X-DNS-Prefetch-Control", sm.policy.DNSPrefetchControl)
		w.Header().Set("Feature-Policy", sm.policy.FeaturePolicy)
		w.Header().Set("Permissions-Policy", sm.policy.PermissionsPolicy)

		next.ServeHTTP(w, r)
	})
}
