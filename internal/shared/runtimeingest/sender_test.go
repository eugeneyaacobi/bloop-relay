package runtimeingest

import (
	"testing"
	"time"

	"bloop-relay/internal/relay/registry"
	relaysession "bloop-relay/internal/relay/session"
)

func TestBuildSnapshotIncludesTunnelAndStaleSessionSignals(t *testing.T) {
	now := time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC)
	snapshot := BuildSnapshot(
		"bloop-relay",
		now,
		[]relaysession.ClientSessionSnapshot{{ID: "acct_a", Name: "client-a", LastSeenAt: now.Add(-2 * time.Minute)}},
		[]registry.TunnelSnapshot{{Name: "api", Hostname: "api.bloop.to", Access: "token_protected", SessionID: "acct_a", Active: false}},
	)

	if len(snapshot.Tunnels) != 1 {
		t.Fatalf("expected one tunnel in snapshot, got %d", len(snapshot.Tunnels))
	}
	if snapshot.Tunnels[0].Status != "down" || !snapshot.Tunnels[0].Degraded {
		t.Fatalf("expected inactive tunnel to be degraded/down, got %+v", snapshot.Tunnels[0])
	}
	if len(snapshot.Events) < 2 {
		t.Fatalf("expected tunnel degraded + stale session events, got %+v", snapshot.Events)
	}
}
