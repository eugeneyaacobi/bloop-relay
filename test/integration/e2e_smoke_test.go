package integration

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEndToEndPublicTunnelSmoke(t *testing.T) {
	t.Run("GET public tunnel", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+stackEnv.relayAddr+"/hello", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Host = "public.bloop.to"
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request through relay: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		if !strings.Contains(string(body), "echo GET /hello") {
			t.Fatalf("unexpected body: %s", string(body))
		}
	})

	t.Run("POST public tunnel preserves body", func(t *testing.T) {
		payload := []byte("ping=post-body")
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://"+stackEnv.relayAddr+"/submit", bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Host = "public.bloop.to"
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request through relay: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		bodyText := string(body)
		if !strings.Contains(bodyText, "echo POST /submit") {
			t.Fatalf("unexpected body: %s", bodyText)
		}
		if !strings.Contains(bodyText, string(payload)) {
			t.Fatalf("expected echoed payload %q in body: %s", string(payload), bodyText)
		}
	})
}
