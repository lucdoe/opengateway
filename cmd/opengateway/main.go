// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
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

// DefaultConfigPath is read when neither -c nor CONFIG_PATH names a file.
const DefaultConfigPath = "./cmd/opengateway/config.yaml"

// resolveConfigPath returns the config file to read: the -c flag, else
// CONFIG_PATH, else DefaultConfigPath.
func resolveConfigPath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		return env
	}
	return DefaultConfigPath
}

func InitializeServer(deps ServerDependencies, configPath string) (*server.Server, error) {
	cfg, err := deps.loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	for _, warning := range cfg.Warnings() {
		log.Printf("warning: %s", warning)
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

const usage = `Usage:
  opengateway [-c config.yaml]         start the gateway
  opengateway check [-c config.yaml]   validate the config and exit

Without -c, the config is read from CONFIG_PATH, else ` + DefaultConfigPath + `.
`

// run executes the command in args (without the program name) and returns
// the exit code: 0 on success, 1 when the config or the server fails, 2 for
// a usage error.
func run(args []string, deps ServerDependencies, stdout, stderr io.Writer) int {
	command := "serve"
	if len(args) > 0 && args[0] == "check" {
		command, args = "check", args[1:]
	}

	flags := flag.NewFlagSet("opengateway", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprint(stderr, usage) }
	configFlag := flags.String("c", "", "config file")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "unknown argument %q\n\n", flags.Arg(0))
		flags.Usage()
		return 2
	}
	configPath := resolveConfigPath(*configFlag)

	if command == "check" {
		return check(deps.loadConfig, configPath, stdout, stderr)
	}
	return serve(deps, configPath)
}

// check loads and validates the config without starting the gateway, so it
// contacts neither the store nor any upstream.
func check(loadConfig func(string) (*config.Config, error), configPath string, stdout, stderr io.Writer) int {
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for _, warning := range cfg.Warnings() {
		fmt.Fprintln(stderr, "warning:", warning)
	}
	endpoints := 0
	for _, service := range cfg.Services {
		endpoints += len(service.Endpoints)
	}
	fmt.Fprintf(stdout, "%s is valid: %d services, %d endpoints\n", configPath, len(cfg.Services), endpoints)
	return 0
}

func serve(deps ServerDependencies, configPath string) int {
	server, err := InitializeServer(deps, configPath)
	if err != nil {
		log.Printf("Failed to initialize the server: %v", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Listening on %s", server.Addr)
	if err := server.Run(ctx); err != nil {
		log.Printf("Failed to run the server: %v", err)
		return 1
	}
	log.Print("Stopped")
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], defaultDependencies(), os.Stdout, os.Stderr))
}
