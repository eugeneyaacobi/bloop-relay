package client

import (
	"context"
	"fmt"
	"log/slog"

	"bloop-tunnel/internal/auth"
	"bloop-tunnel/internal/config"
	"bloop-tunnel/internal/protocol"
	"bloop-tunnel/internal/transport"
)

type Session struct {
	Config     *config.ClientConfig
	Transport  *transport.Session
	Logger     *slog.Logger
	registered []protocol.TunnelRegistration
}

func Connect(ctx context.Context, cfg *config.ClientConfig, logger *slog.Logger) (*Session, error) {
	token := auth.ResolveClientToken(cfg.AuthToken, cfg.AuthTokenEnv)
	if token == "" {
		return nil, fmt.Errorf("missing client auth token")
	}

	ts, err := transport.Dial(ctx, cfg.RelayURL)
	if err != nil {
		return nil, err
	}

	hello := protocol.Envelope{
		Type: protocol.TypeClientHello,
		Payload: protocol.ClientHello{
			Token: token,
			Name:  "default-client",
		},
	}
	if err := ts.Conn.WriteJSON(hello); err != nil {
		_ = ts.Close()
		return nil, err
	}

	var reply protocol.Envelope
	if err := ts.Conn.ReadJSON(&reply); err != nil {
		_ = ts.Close()
		return nil, err
	}
	if reply.Type == protocol.TypeError {
		_ = ts.Close()
		return nil, fmt.Errorf("relay rejected client")
	}

	logger.Info("connected to relay")
	return &Session{Config: cfg, Transport: ts, Logger: logger}, nil
}
