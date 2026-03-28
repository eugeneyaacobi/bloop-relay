package protocol

type MessageType string

const (
	TypeClientHello     MessageType = "client_hello"
	TypeServerHello     MessageType = "server_hello"
	TypeRegisterTunnels MessageType = "register_tunnels"
	TypeRegisterResult  MessageType = "register_result"
	TypeRequestBegin    MessageType = "request_begin"
	TypeRequestBody     MessageType = "request_body_chunk"
	TypeRequestEnd      MessageType = "request_end"
	TypeResponseBegin   MessageType = "response_begin"
	TypeResponseBody    MessageType = "response_body_chunk"
	TypeResponseEnd     MessageType = "response_end"
	TypeError           MessageType = "error"
	TypePing            MessageType = "ping"
	TypePong            MessageType = "pong"
)

type Envelope struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Payload   any         `json:"payload,omitempty"`
}

type ClientHello struct {
	Token string `json:"token"`
	Name  string `json:"name,omitempty"`
}

type TunnelRegistration struct {
	Name              string `json:"name"`
	Hostname          string `json:"hostname,omitempty"`
	LocalAddr         string `json:"local_addr"`
	Access            string `json:"access"`
	BasicAuthUsername string `json:"basic_auth_username,omitempty"`
	BasicAuthPassword string `json:"basic_auth_password,omitempty"`
	ProtectedToken    string `json:"protected_token,omitempty"`
}

type RegisterTunnels struct {
	Tunnels []TunnelRegistration `json:"tunnels"`
}

type TunnelResult struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname,omitempty"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

type RegisterResult struct {
	Results []TunnelResult `json:"results"`
}

type RequestBegin struct {
	RequestID string              `json:"request_id"`
	Method    string              `json:"method"`
	Host      string              `json:"host"`
	Path      string              `json:"path"`
	Headers   map[string][]string `json:"headers,omitempty"`
}

type RequestBodyChunk struct {
	RequestID string `json:"request_id"`
	Data      []byte `json:"data,omitempty"`
}

type ResponseBegin struct {
	RequestID  string              `json:"request_id"`
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers,omitempty"`
}

type ResponseBodyChunk struct {
	RequestID string `json:"request_id"`
	Data      []byte `json:"data,omitempty"`
}
