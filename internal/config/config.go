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

// Names a service or endpoint uses in its Plugins list.
const (
	PluginLogger    = "logger"
	PluginCORS      = "cors"
	PluginRateLimit = "rate-limit"
	PluginAuth      = "auth"
	PluginCache     = "cache"
)

type QueryParam struct {
	Key   string `yaml:"Key"`
	Value string `yaml:"Value"`
}

type AuthConfig struct {
	ApplyAuth bool   `yaml:"ApplyAuth"`
	Method    string `yaml:"Method"`
	Algorithm string `yaml:"Algorithm"`
	Scope     string `yaml:"Scope"`
	SecretKey string `yaml:"SecretKey"`
}

type Endpoint struct {
	Name        string                 `yaml:"Name"`
	HTTPMethod  string                 `yaml:"HTTPMethod"`
	Path        string                 `yaml:"Path"`
	QueryParams []QueryParam           `yaml:"QueryParams"`
	Auth        AuthConfig             `yaml:"Auth"`
	Body        map[string]interface{} `yaml:"Body"`
	Plugins     []string               `yaml:"Plugins"`
}

type Service struct {
	URL       string     `yaml:"URL"`
	Protocol  string     `yaml:"Protocol"`
	Endpoints []Endpoint `yaml:"Endpoints"`
	Plugins   []string   `yaml:"Plugins"`
	Subpath   string     `yaml:"Subpath"`
}

type LoggerConfig struct {
	FilePath  string `yaml:"FilePath"`
	ErrOutput string `yaml:"ErrOutput"`
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
	// Port the gateway listens on. Defaults to DefaultPort.
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
}

// RedisConfig says where the store used by the cache and rate-limit plugins
// is. It is only connected to when one of those plugins is in use.
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
	Redis    RedisConfig        `yaml:"Redis"`
	Plugins  PluginConfig       `yaml:"Plugins"`
	Services map[string]Service `yaml:"Services"`
}

// ListenAddr returns the address the gateway's HTTP server listens on.
func (c *Config) ListenAddr() string {
	return ":" + strconv.Itoa(c.Server.Port)
}

// Load reads the YAML file at path and returns its configuration. It fails
// if the file cannot be read, is not valid YAML, contains a key the gateway
// does not know, or does not pass Validate.
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

// Parse decodes YAML into a validated configuration. Settings that have an
// environment variable (see env.go) are taken from the environment when it is
// set, and settings left out get their default. See Load.
func Parse(data []byte) (*Config, error) {
	cfg := &Config{}

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
