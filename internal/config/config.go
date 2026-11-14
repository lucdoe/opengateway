// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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

type AuthConfig struct {
	ApplyAuth bool   `yaml:"ApplyAuth"`
	Scope     string `yaml:"Scope"`
}

type Endpoint struct {
	Name       string     `yaml:"Name"`
	HTTPMethod string     `yaml:"HTTPMethod"`
	Path       string     `yaml:"Path"`
	Auth       AuthConfig `yaml:"Auth"`
	Plugins    []string   `yaml:"Plugins"`
}

type Service struct {
	URL       string     `yaml:"URL"`
	Endpoints []Endpoint `yaml:"Endpoints"`
	Plugins   []string   `yaml:"Plugins"`
	Subpath   string     `yaml:"Subpath"`
}

type LoggerConfig struct {
	// FilePath is the file request lines are appended to. Empty means stdout.
	FilePath string `yaml:"FilePath"`
}

type JWTConfig struct {
	SecretKey     string `yaml:"SecretKey"`
	SigningMethod string `yaml:"SigningMethod"`
	Issuer        string `yaml:"Issuer"`
	Audience      string `yaml:"Audience"`
	Scope         string `yaml:"Scope"`
}

type RateLimitConfig struct {
	Limit  int64         `yaml:"Limit"`
	Window time.Duration `yaml:"Window"`
}

type CORSConfig struct {
	Origins string `yaml:"Origins"`
	Methods string `yaml:"Methods"`
	Headers string `yaml:"Headers"`
}

type PluginConfig struct {
	LoggerConfig    `yaml:"LoggerConfig"`
	JWTConfig       `yaml:"JWTConfig"`
	RateLimitConfig `yaml:"RateLimitConfig"`
	CORSConfig      `yaml:"CORSConfig"`
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

// Parse decodes YAML into a validated configuration. Environment variables
// (see env.go) override the document, and missing settings get defaults.
func Parse(data []byte) (*Config, error) {
	cfg := &Config{}

	if err := rejectRetiredKeys(data); err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	// An empty document decodes to io.EOF; Validate reports what is missing.
	if err := decoder.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
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
