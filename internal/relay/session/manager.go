package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientSession struct {
	ID          string
	Name        string
	TokenName   string
	Conn        *websocket.Conn
	ConnectedAt time.Time
	LastSeenAt  time.Time
}

type ClientSessionSnapshot struct {
	ID          string
	Name        string
	TokenName   string
	ConnectedAt time.Time
	LastSeenAt  time.Time
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*ClientSession
}

func NewManager() *Manager {
	return &Manager{sessions: make(map[string]*ClientSession)}
}

func (m *Manager) Put(s *ClientSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID] = s
}

func (m *Manager) Get(id string) (*ClientSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

func (m *Manager) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok && s.Conn != nil {
		_ = s.Conn.Close()
	}
	delete(m.sessions, id)
}

func (m *Manager) CloseAny() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.sessions {
		if s != nil && s.Conn != nil {
			return s.Conn.Close()
		}
	}
	return fmt.Errorf("no active sessions")
}

func (m *Manager) Snapshot() []ClientSessionSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ClientSessionSnapshot, 0, len(m.sessions))
	for _, s := range m.sessions {
		if s == nil {
			continue
		}
		out = append(out, ClientSessionSnapshot{ID: s.ID, Name: s.Name, TokenName: s.TokenName, ConnectedAt: s.ConnectedAt, LastSeenAt: s.LastSeenAt})
	}
	return out
}
