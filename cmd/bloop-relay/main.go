package main

import (
	"flag"
	"fmt"
	"context"
	"net/http"
	"os"
	"time"

	"bloop-tunnel/internal/config"
	"bloop-tunnel/internal/logging"
	"bloop-tunnel/internal/relay"
	"bloop-tunnel/internal/relay/registry"
	"bloop-tunnel/internal/relay/routing"
	"bloop-tunnel/internal/relay/session"
	"bloop-tunnel/internal/runtimeingest"
	"bloop-tunnel/pkg/version"
)

func main() {
	configPath := flag.String("config", "deploy/examples/relay.example.yaml", "Path to relay config")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("bloop-relay %s (%s) %s\n", version.Version, version.Commit, version.Date)
		return
	}

	cfg, err := config.LoadRelayConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load relay config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.Logging.Level)
	reg := registry.New()
	mgr := session.NewManager()
	router := routing.New(reg)
	proxy := relay.NewRequestProxy(mgr, logger)
	if cfg.RuntimeIngest.Enabled {
		sender := runtimeingest.NewSender(runtimeingest.Config{Enabled: cfg.RuntimeIngest.Enabled, EndpointURL: cfg.RuntimeIngest.EndpointURL, Secret: cfg.RuntimeIngest.Secret, BearerToken: cfg.RuntimeIngest.BearerToken, Interval: time.Duration(cfg.RuntimeIngest.IntervalSec) * time.Second, Source: "bloop-relay"}, nil, logger, mgr, reg)
		go sender.Start(context.Background())
	}
	sessionHandler := session.NewHandler(cfg, mgr, reg, logger)
	sessionHandler.OnEnvelope = proxy.HandleEnvelope
	httpHandler := &relay.HTTPHandler{Router: router, Proxy: proxy}

	mux := http.NewServeMux()
	mux.Handle("/connect", sessionHandler)
	mux.Handle("/", httpHandler)

	logger.Info("relay starting", "domain", cfg.Domain, "listen_addr", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "relay server failed: %v\n", err)
		os.Exit(1)
	}
}
