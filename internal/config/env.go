// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variables that override the config file when set and non-empty.
const (
	EnvPort          = "PORT"
	EnvStoreType     = "STORE_TYPE"
	EnvRedisHost     = "REDIS_HOST"
	EnvRedisPort     = "REDIS_PORT"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvJWTSecret     = "JWT_SECRET"
)

// Values used when neither the config file nor the environment sets one.
const (
	DefaultPort      = 4000
	DefaultStoreType = StoreMemory
	DefaultRedisHost = "localhost"
	DefaultRedisPort = 6379

	DefaultReadHeaderTimeout = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultShutdownTimeout   = 15 * time.Second
	DefaultUpstreamTimeout   = 30 * time.Second

	DefaultCircuitBreakerConsecutiveFailures = 5
	DefaultCircuitBreakerOpenTimeout         = 30 * time.Second
	DefaultCircuitBreakerHalfOpenRequests    = 1
)

func (c *Config) applyEnvironment() error {
	if err := envInt(EnvPort, &c.Server.Port); err != nil {
		return err
	}
	if err := envInt(EnvRedisPort, &c.Store.Redis.Port); err != nil {
		return err
	}
	envString(EnvStoreType, &c.Store.Type)
	envString(EnvRedisHost, &c.Store.Redis.Host)
	envString(EnvRedisPassword, &c.Store.Redis.Password)
	envString(EnvJWTSecret, &c.Plugins.JWTConfig.SecretKey)
	return nil
}

func (c *Config) applyDefaults() {
	orDefault(&c.Server.Port, DefaultPort)
	orDefault(&c.Server.ReadHeaderTimeout, DefaultReadHeaderTimeout)
	orDefault(&c.Server.IdleTimeout, DefaultIdleTimeout)
	orDefault(&c.Server.ShutdownTimeout, DefaultShutdownTimeout)
	orDefault(&c.Server.UpstreamTimeout, DefaultUpstreamTimeout)
	orDefault(&c.Server.CircuitBreaker.ConsecutiveFailures, DefaultCircuitBreakerConsecutiveFailures)
	orDefault(&c.Server.CircuitBreaker.OpenTimeout, DefaultCircuitBreakerOpenTimeout)
	orDefault(&c.Server.CircuitBreaker.HalfOpenRequests, DefaultCircuitBreakerHalfOpenRequests)
	c.Store.Type = strings.ToLower(strings.TrimSpace(c.Store.Type))
	orDefault(&c.Store.Type, DefaultStoreType)
	orDefault(&c.Store.Redis.Host, DefaultRedisHost)
	orDefault(&c.Store.Redis.Port, DefaultRedisPort)
}

// orDefault sets *field to value if it holds the zero value, which means the
// config left it out.
func orDefault[T comparable](field *T, value T) {
	var zero T
	if *field == zero {
		*field = value
	}
}

func envString(name string, target *string) {
	if value := os.Getenv(name); value != "" {
		*target = value
	}
}

func envInt(name string, target *int) error {
	value := os.Getenv(name)
	if value == "" {
		return nil
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("environment variable %s: %q is not a number", name, value)
	}
	*target = number
	return nil
}
