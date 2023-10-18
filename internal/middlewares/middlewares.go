package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone/internal"
	"github.com/lucdoe/capstone/internal/utils"
	"github.com/microcosm-cc/bluemonday"
)

const (
	// SecurityHeaders constants
	ContentSecurityPolicy = "default-src 'self'; frame-ancestors 'none'; script-src 'self'; object-src 'none'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;"
	ContentTypeOptions    = "nosniff"
	FrameOptions          = "DENY"
	XSSProtection         = "1; mode=block"
	StrictTransport       = "max-age=63072000; includeSubDomains"
	ReferrerPolicy        = "strict-origin-when-cross-origin"
	DNSPrefetchControl    = "off"
	FeaturePolicy         = "geolocation 'none'; midi 'none'; notifications 'none'; push 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; speaker 'none'; vibrate 'none'; fullscreen 'self'; payment 'none';"
	PermissionsPolicy     = "geolocation=(), midi=(), notifications=(), push=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), payment=()"

	// CORS constants
	AllowedOrigin = "http://localhost:3000"
	MaxAgeCORS    = 12 * time.Hour
)

func InitilizeMiddlewares(r *gin.Engine) {
	r.Use(gin.Recovery())

	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Use(LogRequest)
	r.Use(bodySizeLimit)
	r.Use(sanitizeBody)
	r.Use(securityHeaders)
	r.Use(RateLimit)
	r.Use(setCORS)
}

func securityHeaders(c *gin.Context) {
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

func setCORS(c *gin.Context) {
	cors.New(cors.Config{
		AllowOrigins:     []string{AllowedOrigin},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET"},
		AllowHeaders:     []string{"Origin, X-CSRF-Token, Cache-Control"},
		AllowCredentials: true,
		MaxAge:           MaxAgeCORS,
	})
}

func bodySizeLimit(c *gin.Context) {
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	utils.ResetRequestBody(c, bodyBytes)

	size := int64(len(bodyBytes))
	if size > 1<<20 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too big"})
		c.Abort()
		return
	}

	c.Next()
}

func sanitizeBody(c *gin.Context) {
	p := bluemonday.UGCPolicy()

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	utils.ResetRequestBody(c, p.SanitizeBytes(bodyBytes))

	c.Next()
}

func Proxy(serviceName string, endpoint internal.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Executing Proxy middleware")
		c.Request.URL.Host = serviceName
		c.Request.URL.Scheme = "http"
		c.Request.URL.Path = endpoint.Path
	}
}

func ValidateRequest(allowedJSON []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read request body"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		decodedBody := html.UnescapeString(string(body))
		var requestBody map[string]interface{}
		err = json.Unmarshal([]byte(decodedBody), &requestBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if err := checkAllowedFields(requestBody, allowedJSON); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
	}
}

func checkAllowedFields(requestBody map[string]interface{}, allowedJSON []string) error {
	for key := range requestBody {
		if !utils.Contains(allowedJSON, key) {
			return fmt.Errorf("Invalid JSON structure: field %s is not allowed", key)
		}
	}
	return nil
}
