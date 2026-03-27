package integration

import "testing"

func TestClientReconnectRecovery(t *testing.T) {
	assertEventuallyReachable(t, "GET", "public.bloop.to", "/hello", nil, nil, 200, "echo GET /hello")

	if err := stackEnv.disconnectClientTransport(); err != nil {
		t.Fatalf("disconnect client transport: %v", err)
	}
	if err := stackEnv.waitForClientLog("reconnected and re-registered tunnels"); err != nil {
		t.Fatalf("wait for reconnect log: %v", err)
	}

	assertEventuallyReachable(t, "GET", "public.bloop.to", "/hello", nil, nil, 200, "echo GET /hello")
}
