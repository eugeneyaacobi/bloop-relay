package client

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"bloop-tunnel/internal/protocol"
)

func (s *Session) Run() error {
	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			_ = s.Transport.Ping()
		}
	}()

	for {
		var env protocol.Envelope
		if err := s.Transport.Conn.ReadJSON(&env); err != nil {
			return err
		}

		switch env.Type {
		case protocol.TypeRequestBegin:
			if err := s.handleRequest(env); err != nil {
				_ = s.Transport.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, RequestID: env.RequestID, Payload: map[string]string{"error": err.Error()}})
			}
		case protocol.TypePing:
			_ = s.Transport.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypePong})
		}
	}
}

func (s *Session) handleRequest(env protocol.Envelope) error {
	raw, err := protocol.Encode(env.Payload)
	if err != nil {
		return err
	}
	var begin protocol.RequestBegin
	if err := protocol.Decode(raw, &begin); err != nil {
		return err
	}

	if s.Logger != nil {
		s.Logger.Debug("received forwarded request", "request_id", begin.RequestID, "host", begin.Host, "method", begin.Method, "path", begin.Path)
	}

	var body []byte
	for {
		var next protocol.Envelope
		if err := s.Transport.Conn.ReadJSON(&next); err != nil {
			return err
		}
		switch next.Type {
		case protocol.TypeRequestBody:
			rawChunk, err := protocol.Encode(next.Payload)
			if err != nil {
				return err
			}
			var chunk protocol.RequestBodyChunk
			if err := protocol.Decode(rawChunk, &chunk); err != nil {
				return err
			}
			body = append(body, chunk.Data...)
		case protocol.TypeRequestEnd:
			return s.forwardToLocal(begin, body)
		default:
			return nil
		}
	}
}

func (s *Session) forwardToLocal(begin protocol.RequestBegin, body []byte) error {
	tunnel, ok := s.lookupTunnel(begin.Host)
	if !ok {
		if s.Logger != nil {
			s.Logger.Debug("forwarded request host not registered", "request_id", begin.RequestID, "host", begin.Host)
		}
		return s.Transport.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeError, RequestID: begin.RequestID, Payload: map[string]string{"error": "unknown tunnel host"}})
	}

	targetURL := "http://" + tunnel.LocalAddr + begin.Path
	if s.Logger != nil {
		s.Logger.Debug("forwarding request to local service", "request_id", begin.RequestID, "target_url", targetURL)
	}
	req, err := http.NewRequest(begin.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for key, values := range begin.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Debug("local service request failed", "request_id", begin.RequestID, "target_url", targetURL, "error", err)
		}
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if s.Logger != nil {
		s.Logger.Debug("local service responded", "request_id", begin.RequestID, "status_code", resp.StatusCode, "body_bytes", len(respBody))
	}

	if err := s.Transport.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeResponseBegin, RequestID: begin.RequestID, Payload: protocol.ResponseBegin{
		RequestID:  begin.RequestID,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}}); err != nil {
		return err
	}
	if len(respBody) > 0 {
		if err := s.Transport.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeResponseBody, RequestID: begin.RequestID, Payload: protocol.ResponseBodyChunk{RequestID: begin.RequestID, Data: respBody}}); err != nil {
			return err
		}
	}
	return s.Transport.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeResponseEnd, RequestID: begin.RequestID})
}

func (s *Session) lookupTunnel(host string) (protocol.TunnelRegistration, bool) {
	for _, t := range s.registered {
		if t.Hostname == host {
			return t, true
		}
	}
	return protocol.TunnelRegistration{}, false
}
