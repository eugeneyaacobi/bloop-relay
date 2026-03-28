package relay

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"bloop-relay/internal/shared/protocol"
	"bloop-relay/internal/relay/session"
)

type pendingResponse struct {
	statusCode int
	headers    http.Header
	body       bytes.Buffer
	done       chan struct{}
	err        error
	once       sync.Once
}

type RequestProxy struct {
	Sessions *session.Manager
	Logger   *slog.Logger
	pending  sync.Map
}

func NewRequestProxy(sessions *session.Manager, logger *slog.Logger) *RequestProxy {
	return &RequestProxy{Sessions: sessions, Logger: logger}
}

func (p *RequestProxy) ForwardHTTP(w http.ResponseWriter, r *http.Request, sessionID string) error {
	s, ok := p.Sessions.Get(sessionID)
	if !ok {
		return fmt.Errorf("session not found")
	}

	requestID := randomRequestID()
	if p.Logger != nil {
		p.Logger.Debug("forwarding request to client", "request_id", requestID, "session_id", sessionID, "host", r.Host, "method", r.Method, "path", r.URL.RequestURI())
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	pending := &pendingResponse{done: make(chan struct{}), headers: make(http.Header)}
	p.pending.Store(requestID, pending)
	defer p.pending.Delete(requestID)

	if err := s.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeRequestBegin, RequestID: requestID, Payload: protocol.RequestBegin{
		RequestID: requestID,
		Method:    r.Method,
		Host:      r.Host,
		Path:      r.URL.RequestURI(),
		Headers:   r.Header,
	}}); err != nil {
		return err
	}

	if len(body) > 0 {
		if err := s.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeRequestBody, RequestID: requestID, Payload: protocol.RequestBodyChunk{RequestID: requestID, Data: body}}); err != nil {
			return err
		}
	}

	if err := s.Conn.WriteJSON(protocol.Envelope{Type: protocol.TypeRequestEnd, RequestID: requestID}); err != nil {
		return err
	}

	select {
	case <-pending.done:
		if pending.err != nil {
			if p.Logger != nil {
				p.Logger.Debug("client responded with error", "request_id", requestID, "error", pending.err)
			}
			return pending.err
		}
		for key, values := range pending.headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		if pending.statusCode == 0 {
			pending.statusCode = http.StatusBadGateway
		}
		if p.Logger != nil {
			p.Logger.Debug("forwarded response from client", "request_id", requestID, "status_code", pending.statusCode, "body_bytes", pending.body.Len())
		}
		w.WriteHeader(pending.statusCode)
		_, _ = w.Write(pending.body.Bytes())
		return nil
	case <-time.After(30 * time.Second):
		if p.Logger != nil {
			p.Logger.Debug("timed out waiting for client response", "request_id", requestID)
		}
		return fmt.Errorf("timed out waiting for client response")
	}
}

func (p *RequestProxy) HandleEnvelope(env protocol.Envelope) {
	if p.Logger != nil {
		p.Logger.Debug("relay received client envelope", "request_id", env.RequestID, "type", env.Type)
	}
	value, ok := p.pending.Load(env.RequestID)
	if !ok {
		return
	}
	pending := value.(*pendingResponse)

	switch env.Type {
	case protocol.TypeResponseBegin:
		raw, _ := protocol.Encode(env.Payload)
		var payload protocol.ResponseBegin
		if err := protocol.Decode(raw, &payload); err != nil {
			pending.err = err
			pending.once.Do(func() { close(pending.done) })
			return
		}
		pending.statusCode = payload.StatusCode
		for key, values := range payload.Headers {
			for _, value := range values {
				pending.headers.Add(key, value)
			}
		}
	case protocol.TypeResponseBody:
		raw, _ := protocol.Encode(env.Payload)
		var payload protocol.ResponseBodyChunk
		if err := protocol.Decode(raw, &payload); err != nil {
			pending.err = err
			pending.once.Do(func() { close(pending.done) })
			return
		}
		_, _ = pending.body.Write(payload.Data)
	case protocol.TypeResponseEnd:
		pending.once.Do(func() { close(pending.done) })
	case protocol.TypeError:
		pending.err = fmt.Errorf("client returned error")
		pending.once.Do(func() { close(pending.done) })
	}
}

func randomRequestID() string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
