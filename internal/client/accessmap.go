package client

import "bloop-tunnel/internal/config"

func (s *Session) lookupTunnelConfig(host string) (config.TunnelConfig, bool) {
	for _, t := range s.Config.Tunnels {
		if t.Hostname == host {
			return t, true
		}
	}
	return config.TunnelConfig{}, false
}
