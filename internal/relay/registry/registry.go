package registry

import (
	"errors"
	"sync"
)

var ErrHostnameInUse = errors.New("hostname already in use")

type Tunnel struct {
	Name              string
	Hostname          string
	LocalAddr         string
	Access            string
	SessionID         string
	Active            bool
	BasicAuthUsername string
	BasicAuthPassword string
	ProtectedToken    string
}

type TunnelSnapshot struct {
	Name      string
	Hostname  string
	LocalAddr string
	Access    string
	SessionID string
	Active    bool
}

type Registry struct {
	mu      sync.RWMutex
	byHost  map[string]Tunnel
	byOwner map[string][]string
}

func New() *Registry {
	return &Registry{
		byHost:  make(map[string]Tunnel),
		byOwner: make(map[string][]string),
	}
}

func (r *Registry) Register(t Tunnel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.byHost[t.Hostname]; ok && existing.SessionID != t.SessionID {
		return ErrHostnameInUse
	}

	r.byHost[t.Hostname] = t
	r.byOwner[t.SessionID] = appendUnique(r.byOwner[t.SessionID], t.Hostname)
	return nil
}

func (r *Registry) Lookup(hostname string) (Tunnel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.byHost[hostname]
	return t, ok
}

func (r *Registry) ReleaseSession(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hosts := r.byOwner[sessionID]
	for _, host := range hosts {
		delete(r.byHost, host)
	}
	delete(r.byOwner, sessionID)
}

func (r *Registry) Snapshot() []TunnelSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]TunnelSnapshot, 0, len(r.byHost))
	for _, t := range r.byHost {
		out = append(out, TunnelSnapshot{Name: t.Name, Hostname: t.Hostname, LocalAddr: t.LocalAddr, Access: t.Access, SessionID: t.SessionID, Active: t.Active})
	}
	return out
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}
