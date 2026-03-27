package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"bloop-tunnel/internal/auth"
	"bloop-tunnel/internal/client"
	"bloop-tunnel/internal/config"
	"bloop-tunnel/internal/logging"
	"bloop-tunnel/pkg/version"
)

func main() {
	configPath := flag.String("config", "deploy/examples/client.example.yaml", "Path to client config")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("bloop-client %s (%s) %s\n", version.Version, version.Commit, version.Date)
		return
	}

	cfg, err := config.LoadClientConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load client config: %v\n", err)
		os.Exit(1)
	}

	token := auth.ResolveClientToken(cfg.AuthToken, cfg.AuthTokenEnv)
	logger := logging.New(cfg.Logging.Level)
	logger.Info("client starting", "relay_url", cfg.RelayURL, "tunnel_count", len(cfg.Tunnels), "has_token", token != "")

	if cfg.ControlPlaneURL != "" {
		enrollmentToken := cfg.EnrollmentToken
		if enrollmentToken == "" && cfg.EnrollmentTokenEnv != "" {
			enrollmentToken = os.Getenv(cfg.EnrollmentTokenEnv)
		}
		if enrollmentToken != "" {
			installationID, ingestToken, err := client.EnrollRuntime(context.Background(), cfg.ControlPlaneURL, enrollmentToken, "default-client")
			if err != nil {
				fmt.Fprintf(os.Stderr, "enroll runtime: %v\n", err)
				os.Exit(1)
			}
			logger.Info("runtime enrollment succeeded", "installation_id", installationID)
			os.Setenv("BLOOP_RUNTIME_INSTALLATION_ID", installationID)
			os.Setenv("BLOOP_RUNTIME_INGEST_BEARER", ingestToken)
		}
	}

	session, err := client.Connect(context.Background(), cfg, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to relay: %v\n", err)
		os.Exit(1)
	}
	defer session.Transport.Close()

	if err := validateProtectedTunnels(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "invalid tunnel config: %v\n", err)
		os.Exit(1)
	}

	if err := session.RegisterTunnels(); err != nil {
		fmt.Fprintf(os.Stderr, "register tunnels: %v\n", err)
		os.Exit(1)
	}

	logger.Info("client registered tunnels successfully")
	registered := make([]string, 0, len(cfg.Tunnels))
	for _, t := range cfg.Tunnels {
		registered = append(registered, t.Name)
	}
	go func() {
		if err := client.StartRuntimeIngestLoop(context.Background(), cfg, registered); err != nil {
			logger.Error("runtime ingest loop failed", "error", err.Error())
		}
	}()
	if err := session.RunWithReconnect(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "client session ended: %v\n", err)
		os.Exit(1)
	}
}

func validateProtectedTunnels(cfg *config.ClientConfig) error {
	for _, t := range cfg.Tunnels {
		switch t.Access {
		case "basic_auth":
			if t.BasicAuth.Username == "" {
				return fmt.Errorf("tunnel %q missing basic auth username", t.Name)
			}
			if t.BasicAuth.Password == "" && t.BasicAuth.PasswordEnv == "" {
				return fmt.Errorf("tunnel %q missing basic auth password or password_env", t.Name)
			}
		case "token_protected":
			if t.Token == "" && t.TokenEnv == "" {
				return fmt.Errorf("tunnel %q missing token or token_env", t.Name)
			}
		case "public", "":
		default:
			return errors.New("unsupported access mode for tunnel " + t.Name)
		}
	}
	return nil
}
