package transport

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type Session struct {
	Conn *websocket.Conn
}

func Dial(ctx context.Context, url string) (*Session, error) {
	d := websocket.Dialer{}
	conn, _, err := d.DialContext(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(10 << 20)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	return &Session{Conn: conn}, nil
}

func (s *Session) SetReadDeadline(timeout time.Duration) error {
	return s.Conn.SetReadDeadline(time.Now().Add(timeout))
}

func (s *Session) Ping() error {
	return s.Conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
}

func (s *Session) Close() error {
	return s.Conn.Close()
}
