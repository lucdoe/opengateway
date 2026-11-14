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
	"fmt"
	"os"
	"strconv"
	"time"
)

// Environment variables that override the config file. A variable that is
// unset or empty leaves the value from the file in place.
const (
	EnvPort          = "PORT"
	EnvRedisHost     = "REDIS_HOST"
	EnvRedisPort     = "REDIS_PORT"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvJWTSecret     = "JWT_SECRET"
)

// Values used when neither the config file nor the environment sets one.
const (
	DefaultPort      = 4000
	DefaultRedisHost = "localhost"
	DefaultRedisPort = 6379

	DefaultReadHeaderTimeout = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultShutdownTimeout   = 15 * time.Second
	DefaultUpstreamTimeout   = 30 * time.Second
)

func (c *Config) applyEnvironment() error {
	if err := envInt(EnvPort, &c.Server.Port); err != nil {
		return err
	}
	if err := envInt(EnvRedisPort, &c.Redis.Port); err != nil {
		return err
	}
	envString(EnvRedisHost, &c.Redis.Host)
	envString(EnvRedisPassword, &c.Redis.Password)
	envString(EnvJWTSecret, &c.Plugins.JWTConfig.SecretKey)
	return nil
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = DefaultPort
	}
	if c.Server.ReadHeaderTimeout == 0 {
		c.Server.ReadHeaderTimeout = DefaultReadHeaderTimeout
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = DefaultIdleTimeout
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = DefaultShutdownTimeout
	}
	if c.Server.UpstreamTimeout == 0 {
		c.Server.UpstreamTimeout = DefaultUpstreamTimeout
	}
	if c.Redis.Host == "" {
		c.Redis.Host = DefaultRedisHost
	}
	if c.Redis.Port == 0 {
		c.Redis.Port = DefaultRedisPort
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
