// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/lucdoe/opengateway/internal/plugins/auth"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/lucdoe/opengateway/internal/server"
	"github.com/lucdoe/opengateway/internal/store"
	"github.com/lucdoe/opengateway/internal/store/memory"
	"github.com/lucdoe/opengateway/internal/store/redisstore"
)

// ServerDependencies holds what reaches outside the process, so tests can
// replace it.
type ServerDependencies struct {
	loadConfig func(path string) (*config.Config, error)
	router     *mux.Router
	newProxy   server.ProxyFactory
	// newStore is only called when the cache or rate-limit plugin is in use.
	newStore storeFactory
}

func InitializeServer(deps ServerDependencies) (*server.Server, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./cmd/opengateway/config.yaml"
	}
	cfg, err := deps.loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	plugins, closers, err := buildPlugins(cfg, deps.newStore)
	if err != nil {
		return nil, err
	}

	s, err := server.NewServer(cfg, deps.router, deps.newProxy, plugins, requireScope)
	if err != nil {
		for _, closer := range closers {
			closer.Close()
		}
		return nil, err
	}
	s.Closers = closers
	return s, nil
}

// openStore opens the store cfg names. Validation has already rejected
// unknown types.
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

func requireScope(scope string) server.Middleware {
	return auth.RequireScope(scope)
}

// defaultDependencies are the real ones main runs with.
func defaultDependencies() ServerDependencies {
	return ServerDependencies{
		loadConfig: config.Load,
		router:     mux.NewRouter(),
		newProxy:   proxy.NewHandler,
		newStore:   openStore,
	}
}

func main() {
	server, err := InitializeServer(defaultDependencies())
	if err != nil {
		log.Fatalf("Failed to initialize the server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Listening on %s", server.Addr)
	if err := server.Run(ctx); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
	log.Print("Stopped")
}
