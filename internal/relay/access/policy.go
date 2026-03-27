package access

import "bloop-tunnel/internal/config"

type Mode string

const (
	ModePublic         Mode = "public"
	ModeBasicAuth      Mode = "basic_auth"
	ModeTokenProtected Mode = "token_protected"
)

func ResolveMode(value string) Mode {
	switch Mode(value) {
	case ModeBasicAuth:
		return ModeBasicAuth
	case ModeTokenProtected:
		return ModeTokenProtected
	default:
		return ModePublic
	}
}

func FindTunnelConfig(tunnels []config.TunnelConfig, hostname, name string) (config.TunnelConfig, bool) {
	for _, t := range tunnels {
		if t.Hostname == hostname || (hostname == "" && t.Name == name) {
			return t, true
		}
	}
	return config.TunnelConfig{}, false
}
