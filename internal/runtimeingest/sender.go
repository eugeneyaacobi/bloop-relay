package runtimeingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"bloop-tunnel/internal/relay/registry"
	relaysession "bloop-tunnel/internal/relay/session"
)

type Config struct {
	Enabled        bool
	EndpointURL    string
	Secret         string
	BearerToken    string
	InstallationID string
	Interval       time.Duration
	Source         string
}

type Sender struct {
	cfg      Config
	client   *http.Client
	logger   *slog.Logger
	sessions *relaysession.Manager
	registry *registry.Registry
}

type SnapshotPayload struct {
	Source     string               `json:"source"`
	CapturedAt string               `json:"capturedAt"`
	Tunnels    []SnapshotTunnel     `json:"tunnels"`
	Events     []SnapshotEvent      `json:"events"`
}

type SnapshotTunnel struct {
	TunnelID   string `json:"tunnelId"`
	AccountID  string `json:"accountId"`
	Hostname   string `json:"hostname"`
	AccessMode string `json:"accessMode"`
	Status     string `json:"status"`
	Degraded   bool   `json:"degraded"`
	ObservedAt string `json:"observedAt"`
}

type SnapshotEvent struct {
	ID         string `json:"id"`
	AccountID  string `json:"accountId,omitempty"`
	TunnelID   string `json:"tunnelId,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Kind       string `json:"kind"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	OccurredAt string `json:"occurredAt"`
}

func NewSender(cfg Config, client *http.Client, logger *slog.Logger, sessions *relaysession.Manager, reg *registry.Registry) *Sender {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.Source == "" {
		cfg.Source = "bloop-relay"
	}
	return &Sender{cfg: cfg, client: client, logger: logger, sessions: sessions, registry: reg}
}

func (s *Sender) Start(ctx context.Context) {
	if s == nil || !s.cfg.Enabled || strings.TrimSpace(s.cfg.EndpointURL) == "" || (strings.TrimSpace(s.cfg.Secret) == "" && strings.TrimSpace(s.cfg.BearerToken) == "") {
		return
	}
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()
	_ = s.send(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.send(ctx)
		}
	}
}

func (s *Sender) send(ctx context.Context) error {
	payload := BuildSnapshot(s.cfg.Source, time.Now().UTC(), s.sessions.Snapshot(), s.registry.Snapshot())
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.EndpointURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(s.cfg.BearerToken) != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.BearerToken)
	} else {
		req.Header.Set("X-Bloop-Runtime-Secret", s.cfg.Secret)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("runtime ingest send failed", "error", err.Error())
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("runtime ingest status %d", resp.StatusCode)
	}
	if s.logger != nil {
		s.logger.Info("runtime ingest snapshot sent", "tunnels", len(payload.Tunnels), "events", len(payload.Events))
	}
	return nil
}

func BuildSnapshot(source string, now time.Time, sessions []relaysession.ClientSessionSnapshot, tunnels []registry.TunnelSnapshot) SnapshotPayload {
	payload := SnapshotPayload{Source: source, CapturedAt: now.Format(time.RFC3339)}
	installationID := strings.TrimSpace(os.Getenv("BLOOP_RUNTIME_INSTALLATION_ID"))
	for _, tunnel := range tunnels {
		status := "healthy"
		degraded := false
		if !tunnel.Active {
			status = "down"
			degraded = true
		}
		accountID := tunnel.SessionID
		if installationID != "" {
			accountID = installationID
		}
		payload.Tunnels = append(payload.Tunnels, SnapshotTunnel{
			TunnelID:   tunnel.Name,
			AccountID:  accountID,
			Hostname:   tunnel.Hostname,
			AccessMode: tunnel.Access,
			Status:     status,
			Degraded:   degraded,
			ObservedAt: now.Format(time.RFC3339),
		})
		if degraded {
			payload.Events = append(payload.Events, SnapshotEvent{
				ID:         fmt.Sprintf("evt-%s-%d", tunnel.Name, now.Unix()),
				AccountID:  accountID,
				TunnelID:   tunnel.Name,
				Hostname:   tunnel.Hostname,
				Kind:       "tunnel.degraded",
				Level:      "warn",
				Message:    fmt.Sprintf("Tunnel %s is not active on the relay", tunnel.Hostname),
				OccurredAt: now.Format(time.RFC3339),
			})
		}
	}
	for _, sess := range sessions {
		if now.Sub(sess.LastSeenAt) > time.Minute {
			payload.Events = append(payload.Events, SnapshotEvent{
				ID:         fmt.Sprintf("evt-session-%s-%d", sess.ID, now.Unix()),
				AccountID:  sess.ID,
				Kind:       "session.stale",
				Level:      "warn",
				Message:    fmt.Sprintf("Relay session %s has not been seen recently", sess.Name),
				OccurredAt: now.Format(time.RFC3339),
			})
		}
	}
	return payload
}
