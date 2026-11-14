// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Plugin names, as listed in a Plugins list.
const (
	PluginLogger    = "logger"
	PluginCORS      = "cors"
	PluginRateLimit = "rate-limit"
	PluginAuth      = "auth"
	PluginCache     = "cache"
)

// PluginOrder lists every plugin in the order they run for a request,
// outermost first. Auth runs before cache so a cached response is never
// served to a caller that has not been authenticated. See PluginOrderFor
// for the one exception.
var PluginOrder = []string{PluginLogger, PluginCORS, PluginRateLimit, PluginAuth, PluginCache}

// PluginOrderFor is the order for a route whose rate limit counts by
// rateLimitKey: PluginOrder, except that rate-limit runs right after auth
// when it counts per token subject, so it only ever reads claims auth has
// verified.
func PluginOrderFor(rateLimitKey string) []string {
	if rateLimitKey != RateLimitBySubject {
		return PluginOrder
	}
	return []string{PluginLogger, PluginCORS, PluginAuth, PluginRateLimit, PluginCache}
}

// RateLimitConfig.Key values. A header is named after RateLimitByHeader,
// as in header:CF-Connecting-IP.
const (
	RateLimitByIP      = "ip"
	RateLimitByHeader  = "header:"
	RateLimitBySubject = "jwt:sub"
)

type Endpoint struct {
	Name       string   `yaml:"Name"`
	HTTPMethod string   `yaml:"HTTPMethod"`
	Path       string   `yaml:"Path"`
	Plugins    []string `yaml:"Plugins"`
	// Scope lists space-separated scopes a token needs for this endpoint, on
	// top of Plugins.JWT.Scope. It needs the auth plugin.
	Scope    string   `yaml:"Scope"`
	Settings Settings `yaml:"Settings"`
	// UpstreamPath, when set, is the path sent upstream instead of the
	// gateway path, after the service URL's own path. Its {name}s take the
	// values of Path's, and a final /* what came after Path's prefix.
	UpstreamPath string `yaml:"UpstreamPath"`
}

type Service struct {
	URL       string     `yaml:"URL"`
	Endpoints []Endpoint `yaml:"Endpoints"`
	Plugins   []string   `yaml:"Plugins"`
	Subpath   string     `yaml:"Subpath"`
	Settings  Settings   `yaml:"Settings"`
}

type LoggerConfig struct {
	// FilePath is the file request lines are appended to. Empty means stdout.
	FilePath string `yaml:"FilePath"`
}

// JWTConfig says which tokens the auth plugin accepts. Tokens are verified
// with SecretKey, or with the keys an identity provider publishes at
// JWKSURL; exactly one is set.
type JWTConfig struct {
	SecretKey string `yaml:"SecretKey"`
	JWKSURL   string `yaml:"JWKSURL"`
	// SigningMethod limits tokens to one algorithm. Empty means HS256 with
	// SecretKey, and every algorithm in KeyAlgorithms with JWKSURL.
	SigningMethod string `yaml:"SigningMethod"`
	Issuer        string `yaml:"Issuer"`
	Audience      string `yaml:"Audience"`
	Scope         string `yaml:"Scope"`
}

// Algorithms the auth plugin accepts with a SecretKey, and with a JWKSURL.
// Each list only holds algorithms of its own kind, so a token can never get
// a public key used as an HMAC secret.
var (
	HMACAlgorithms = []string{"HS256", "HS384", "HS512"}
	KeyAlgorithms  = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384", "ES512", "EdDSA"}
)

// Algorithms returns the alg values tokens may be signed with.
func (j JWTConfig) Algorithms() []string {
	switch {
	case j.SigningMethod != "":
		return []string{j.SigningMethod}
	case j.JWKSURL != "":
		return KeyAlgorithms
	}
	return []string{"HS256"}
}

type RateLimitConfig struct {
	Limit  int64         `yaml:"Limit"`
	Window time.Duration `yaml:"Window"`
	// Key is what requests are counted by: RateLimitByIP (the default),
	// a request header, or RateLimitBySubject.
	Key string `yaml:"Key"`
}

type CORSConfig struct {
	Origins string `yaml:"Origins"`
	Methods string `yaml:"Methods"`
	Headers string `yaml:"Headers"`
}

type PluginConfig struct {
	LoggerConfig    `yaml:"Logger"`
	JWTConfig       `yaml:"JWT"`
	RateLimitConfig `yaml:"RateLimit"`
	CORSConfig      `yaml:"CORS"`
}

// ServerConfig holds the settings of the gateway's own HTTP server.
type ServerConfig struct {
	// Port is where the gateway listens. Defaults to DefaultPort.
	Port int `yaml:"Port"`
	// ReadHeaderTimeout is how long a client has to send its request headers.
	ReadHeaderTimeout time.Duration `yaml:"ReadHeaderTimeout"`
	// IdleTimeout is how long an idle keep-alive connection stays open.
	IdleTimeout time.Duration `yaml:"IdleTimeout"`
	// ShutdownTimeout is how long requests in flight get to finish when the
	// gateway is asked to stop.
	ShutdownTimeout time.Duration `yaml:"ShutdownTimeout"`
	// UpstreamTimeout is how long the gateway waits for an upstream to start
	// answering before it gives up with 504 Gateway Timeout.
	UpstreamTimeout time.Duration `yaml:"UpstreamTimeout"`
	// CircuitBreaker sets when the gateway stops sending requests to a
	// failing upstream. Every service has its own breaker.
	CircuitBreaker CircuitBreakerConfig `yaml:"CircuitBreaker"`
}

// CircuitBreakerConfig holds the thresholds of the breaker each service has.
// A failure is an upstream that cannot be reached, times out or answers
// 500, 502, 503 or 504.
type CircuitBreakerConfig struct {
	// ConsecutiveFailures is how many failures in a row open the breaker.
	ConsecutiveFailures int `yaml:"ConsecutiveFailures"`
	// OpenTimeout is how long the breaker stays open before it lets trial
	// requests through again.
	OpenTimeout time.Duration `yaml:"OpenTimeout"`
	// HalfOpenRequests is how many trial requests may run at once after
	// OpenTimeout.
	HalfOpenRequests int `yaml:"HalfOpenRequests"`
}

// Store types for StoreConfig.Type.
const (
	StoreMemory = "memory"
	StoreRedis  = "redis"
)

// StoreConfig picks where the cache and rate-limit plugins keep their state.
// It is only used when one of those plugins is in use.
type StoreConfig struct {
	// Type is StoreMemory (the default) or StoreRedis. Memory is per gateway
	// instance; Redis is shared by every instance using the same server.
	Type string `yaml:"Type"`
	// Redis is only read when Type is StoreRedis.
	Redis RedisConfig `yaml:"Redis"`
}

// RedisConfig locates the Redis server.
type RedisConfig struct {
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	Password string `yaml:"Password"`
}

// Addr returns the host:port address of the Redis server.
func (r RedisConfig) Addr() string {
	return net.JoinHostPort(r.Host, strconv.Itoa(r.Port))
}

type Config struct {
	Server   ServerConfig       `yaml:"Server"`
	Store    StoreConfig        `yaml:"Store"`
	Plugins  PluginConfig       `yaml:"Plugins"`
	Services map[string]Service `yaml:"Services"`
}

// ServiceNames returns the names of the services, sorted, so routes and
// error messages come out in the same order every time.
func (c *Config) ServiceNames() []string {
	names := make([]string, 0, len(c.Services))
	for name := range c.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListenAddr returns the address the gateway's HTTP server listens on.
func (c *Config) ListenAddr() string {
	return ":" + strconv.Itoa(c.Server.Port)
}

// Load reads the YAML file at path and returns its validated configuration.
// See Parse.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("config %s: %w", path, err)
	}
	return cfg, nil
}

// Parse decodes YAML into a validated configuration. ${NAME} placeholders
// are filled from the environment (see placeholders.go), environment
// variables override the document (see env.go), and missing settings get
// defaults.
func Parse(data []byte) (*Config, error) {
	cfg := &Config{}

	if err := rejectRetiredKeys(data); err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if err := rejectUnknownKeys(data); err != nil {
		return nil, err
	}
	if err := fillPlaceholders(&doc); err != nil {
		return nil, err
	}
	// An empty document has no content; Validate reports what is missing.
	if doc.Kind != 0 {
		if err := doc.Decode(cfg); err != nil {
			return nil, err
		}
	}

	if err := cfg.applyEnvironment(); err != nil {
		return nil, err
	}
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// rejectUnknownKeys reports keys the gateway does not read, such as a
// misspelled one, with their lines. It decodes the document as written,
// before placeholders are filled, so it skips the type errors an unfilled
// placeholder like Port: ${PORT} gives; the decode after filling reports
// the real ones.
func rejectUnknownKeys(data []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var typeErr *yaml.TypeError
	if err := decoder.Decode(&Config{}); !errors.As(err, &typeErr) {
		return nil // io.EOF for an empty document; syntax errors are reported before
	}
	var unknown []string
	for _, problem := range typeErr.Errors {
		if strings.Contains(problem, " not found in type ") {
			unknown = append(unknown, problem)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	return &yaml.TypeError{Errors: unknown}
}
