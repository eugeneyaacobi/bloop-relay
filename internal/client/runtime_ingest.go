package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"bloop-tunnel/internal/config"
)

type RuntimeIngestPayload struct {
	CapturedAt string                `json:"capturedAt"`
	Tunnels    []RuntimeIngestTunnel `json:"tunnels"`
	Events     []RuntimeIngestEvent  `json:"events"`
}

type RuntimeIngestTunnel struct {
	TunnelID   string `json:"tunnelId"`
	AccountID  string `json:"accountId,omitempty"`
	Hostname   string `json:"hostname"`
	AccessMode string `json:"accessMode"`
	Status     string `json:"status"`
	Degraded   bool   `json:"degraded"`
	ObservedAt string `json:"observedAt"`
}

type RuntimeIngestEvent struct {
	ID         string `json:"id"`
	AccountID  string `json:"accountId,omitempty"`
	TunnelID   string `json:"tunnelId,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Kind       string `json:"kind"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	OccurredAt string `json:"occurredAt"`
}

func StartRuntimeIngestLoop(ctx context.Context, cfg *config.ClientConfig, registered []string) error {
	controlPlaneURL := strings.TrimRight(strings.TrimSpace(cfg.ControlPlaneURL), "/")
	ingestToken := strings.TrimSpace(os.Getenv("BLOOP_RUNTIME_INGEST_BEARER"))
	installationID := strings.TrimSpace(os.Getenv("BLOOP_RUNTIME_INSTALLATION_ID"))
	if controlPlaneURL == "" || ingestToken == "" {
		return nil
	}
	endpoint := controlPlaneURL + "/api/runtime/ingest/snapshot"
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	if err := sendRuntimeSnapshot(ctx, endpoint, ingestToken, installationID, cfg.Tunnels, registered); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := sendRuntimeSnapshot(ctx, endpoint, ingestToken, installationID, cfg.Tunnels, registered); err != nil {
				return err
			}
		}
	}
}

func sendRuntimeSnapshot(ctx context.Context, endpoint, ingestToken, installationID string, tunnels []config.TunnelConfig, registered []string) error {
	now := time.Now().UTC()
	registeredSet := map[string]bool{}
	for _, name := range registered {
		registeredSet[name] = true
	}
	payload := RuntimeIngestPayload{CapturedAt: now.Format(time.RFC3339)}
	for _, t := range tunnels {
		status := "down"
		degraded := true
		if registeredSet[t.Name] {
			status = "healthy"
			degraded = false
		}
		payload.Tunnels = append(payload.Tunnels, RuntimeIngestTunnel{TunnelID: t.Name, AccountID: installationID, Hostname: t.Hostname, AccessMode: t.Access, Status: status, Degraded: degraded, ObservedAt: now.Format(time.RFC3339)})
		if degraded {
			payload.Events = append(payload.Events, RuntimeIngestEvent{ID: fmt.Sprintf("evt-%s-%d", t.Name, now.Unix()), AccountID: installationID, TunnelID: t.Name, Hostname: t.Hostname, Kind: "tunnel.degraded", Level: "warn", Message: fmt.Sprintf("Tunnel %s not registered", t.Name), OccurredAt: now.Format(time.RFC3339)})
		}
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil { return err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ingestToken)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { return fmt.Errorf("runtime ingest status %d", resp.StatusCode) }
	return nil
}
