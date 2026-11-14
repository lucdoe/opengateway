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
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/lucdoe/opengateway/internal/server"
	"github.com/lucdoe/opengateway/internal/store"
	"github.com/lucdoe/opengateway/internal/store/memory"
	"github.com/lucdoe/opengateway/internal/store/redisstore"
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
	// newStore opens the store shared by the cache and rate-limit plugins.
	// It is only called when one of them is in use.
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

	plugins, closers, err := buildPlugins(cfg, deps.newStore)
	if err != nil {
		return nil, err
	}

	s, err := server.NewServer(cfg, deps.router, deps.newProxy, plugins)
	if err != nil {
		for _, closer := range closers {
			closer.Close()
		}
		return nil, err
	}
	s.Closers = closers
	return s, nil
}

// openStore opens the store that cfg.Type names. The config is validated, so
// the type is a known one.
func openStore(cfg config.StoreConfig) (store.Store, error) {
	switch cfg.Type {
	case config.StoreRedis:
		s, err := redisstore.New(redisstore.Config{Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password})
		if err != nil {
			return nil, err
		}
		return s, nil
	case config.StoreMemory:
		return memory.New(), nil
	default:
		return nil, fmt.Errorf("unknown store type %q", cfg.Type)
	}
}

func main() {
	deps := ServerDependencies{
		configLoader: &DefaultConfigLoader{},
		router:       mux.NewRouter(),
		newProxy:     proxy.NewHandler,
		newStore:     openStore,
	}

	server, err := InitializeServer(deps)
	if err != nil {
		log.Fatalf("Failed to initialize the server: %v", err)
	}

	// SIGINT and SIGTERM ask the gateway to stop: it finishes the requests in
	// flight and then exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Listening on %s", server.Addr)
	if err := server.Run(ctx); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
	log.Print("Stopped")
}
