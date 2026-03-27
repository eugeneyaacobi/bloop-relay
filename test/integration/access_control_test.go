package integration

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestAccessControlModes(t *testing.T) {
	t.Run("basic auth missing returns 401", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+stackEnv.relayAddr+"/hello", nil)
		req.Host = "basic.bloop.to"
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("basic auth good returns 200", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+stackEnv.relayAddr+"/hello", nil)
		req.Host = "basic.bloop.to"
		req.SetBasicAuth("gene", "secretpass")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("token missing returns 401", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+stackEnv.relayAddr+"/hello", nil)
		req.Host = "token.bloop.to"
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 401, got %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("token good returns 200", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+stackEnv.relayAddr+"/hello", nil)
		req.Host = "token.bloop.to"
		req.Header.Set("X-Bloop-Token", "topsecret")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
	})
}
