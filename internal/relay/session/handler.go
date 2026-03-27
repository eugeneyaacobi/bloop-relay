package session

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"bloop-tunnel/internal/auth"
	"bloop-tunnel/internal/config"
	"bloop-tunnel/internal/generator"
	"bloop-tunnel/internal/protocol"
	"bloop-tunnel/internal/relay/registry"

	"github.com/gorilla/websocket"
)

type Handler struct {
	Upgrader   websocket.Upgrader
	Manager    *Manager
	Registry   *registry.Registry
	Config     *config.RelayConfig
	Logger     *slog.Logger
	OnEnvelope func(protocol.Envelope)
}

func NewHandler(cfg *config.RelayConfig, mgr *Manager, reg *registry.Registry, logger *slog.Logger) *Handler {
	return &Handler{
		Upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		Manager:  mgr,
		Registry: reg,
		Config:   cfg,
		Logger:   logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Logger.Error("upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	var env protocol.Envelope
	if err := conn.ReadJSON(&env); err != nil {
		h.Logger.Error("read envelope failed", "error", err)
		return
	}
	if env.Type != protocol.TypeClientHello {
		_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, Payload: map[string]string{"error": "expected client_hello"}})
		return
	}

	raw, err := protocol.Encode(env.Payload)
	if err != nil {
		_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, Payload: map[string]string{"error": "invalid payload"}})
		return
	}
	var hello protocol.ClientHello
	if err := protocol.Decode(raw, &hello); err != nil {
		_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, Payload: map[string]string{"error": "invalid client_hello"}})
		return
	}

	tokenName, ok := validateToken(h.Config.ClientTokens, hello.Token)
	if !ok {
		_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, Payload: map[string]string{"error": "unauthorized"}})
		return
	}

	sessionID := randomID()
	s := &ClientSession{
		ID:          sessionID,
		Name:        hello.Name,
		TokenName:   tokenName,
		Conn:        conn,
		ConnectedAt: time.Now().UTC(),
		LastSeenAt:  time.Now().UTC(),
	}
	h.Manager.Put(s)
	defer func() {
		h.Registry.ReleaseSession(sessionID)
		h.Manager.Delete(sessionID)
	}()

	_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeServerHello, Payload: map[string]string{"session_id": sessionID}})
	h.Logger.Info("client connected", "session_id", sessionID, "name", hello.Name, "token_name", tokenName)

	for {
		var msg protocol.Envelope
		if err := conn.ReadJSON(&msg); err != nil {
			h.Logger.Info("client disconnected", "session_id", sessionID, "error", err)
			return
		}
		s.LastSeenAt = time.Now().UTC()

		switch msg.Type {
		case protocol.TypeRegisterTunnels:
			h.handleRegister(conn, sessionID, msg)
		case protocol.TypePing:
			_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypePong})
		case protocol.TypePong:
			// read deadline already refreshed by transport pong handler on the client side; keep as no-op here
		case protocol.TypeResponseBegin, protocol.TypeResponseBody, protocol.TypeResponseEnd:
			if h.OnEnvelope != nil {
				h.OnEnvelope(msg)
			}
		default:
			_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, Payload: map[string]string{"error": "unsupported message type"}})
		}
	}
}

func (h *Handler) handleRegister(conn *websocket.Conn, sessionID string, msg protocol.Envelope) {
	raw, err := protocol.Encode(msg.Payload)
	if err != nil {
		_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, Payload: map[string]string{"error": "invalid register payload"}})
		return
	}
	var req protocol.RegisterTunnels
	if err := protocol.Decode(raw, &req); err != nil {
		_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, Payload: map[string]string{"error": "invalid register_tunnels"}})
		return
	}

	results := make([]protocol.TunnelResult, 0, len(req.Tunnels))
	for _, t := range req.Tunnels {
		hostname := strings.TrimSpace(t.Hostname)
		if hostname == "" {
			label, err := generator.RandomLabel(h.Config.HostnameGeneration.Length)
			if err != nil {
				results = append(results, protocol.TunnelResult{Name: t.Name, Success: false, Error: "hostname generation failed"})
				continue
			}
			hostname = generator.BuildHostname(label, h.Config.Domain)
		}

		err := h.Registry.Register(registry.Tunnel{
			Name:              t.Name,
			Hostname:          hostname,
			LocalAddr:         t.LocalAddr,
			Access:            t.Access,
			SessionID:         sessionID,
			Active:            true,
			BasicAuthUsername: t.BasicAuthUsername,
			BasicAuthPassword: t.BasicAuthPassword,
			ProtectedToken:    t.ProtectedToken,
		})
		if err != nil {
			results = append(results, protocol.TunnelResult{Name: t.Name, Hostname: hostname, Success: false, Error: err.Error()})
			continue
		}
		results = append(results, protocol.TunnelResult{Name: t.Name, Hostname: hostname, Success: true})
	}

	_ = conn.WriteJSON(protocol.Envelope{Type: protocol.TypeRegisterResult, Payload: protocol.RegisterResult{Results: results}})
}

func validateToken(tokens []config.RelayToken, provided string) (string, bool) {
	for _, t := range tokens {
		resolved := auth.ResolveRelayToken(t)
		if resolved != "" && resolved == provided {
			return t.Name, true
		}
	}
	return "", false
}

func randomID() string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
