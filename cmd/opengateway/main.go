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

package main

import (
	"log"
	"os"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/plugins/cache"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/lucdoe/opengateway/internal/server"
)

type configLoader interface {
	LoadConfig(path string) (*config.Config, error)
}

type DefaultConfigLoader struct{}

func (d *DefaultConfigLoader) LoadConfig(path string) (*config.Config, error) {
	return config.Load(path)
}

// ServerDependencies are the parts of the gateway that talk to the outside
// world, so tests can replace them.
type ServerDependencies struct {
	configLoader configLoader
	router       *mux.Router
	newProxy     server.ProxyFactory
	// newStore connects to the store shared by the cache and rate-limit
	// plugins. It is only called when one of them is in use.
	newStore storeFactory
}

func InitializeServer(deps ServerDependencies) (*server.Server, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./cmd/opengateway/config.yaml"
	}
	cfg, err := deps.configLoader.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	plugins, err := buildPlugins(cfg, deps.newStore)
	if err != nil {
		return nil, err
	}

	return server.NewServer(cfg, deps.router, deps.newProxy, plugins)
}

func connectRedis(cfg config.RedisConfig) (cache.Cache, error) {
	store, err := cache.NewRedisCache(cache.CacheConfig{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

func main() {
	deps := ServerDependencies{
		configLoader: &DefaultConfigLoader{},
		router:       mux.NewRouter(),
		newProxy:     proxy.NewHandler,
		newStore:     connectRedis,
	}

	server, err := InitializeServer(deps)
	if err != nil {
		log.Fatalf("Failed to initialize the server: %v", err)
	}

	log.Printf("Listening on %s", server.Addr)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}
