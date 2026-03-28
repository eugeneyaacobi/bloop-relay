package access

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"bloop-relay/internal/relay/registry"
)

func EnforceTunnel(r *http.Request, tunnel registry.Tunnel) (int, string, bool) {
	switch ResolveMode(tunnel.Access) {
	case ModeBasicAuth:
		return enforceTunnelBasicAuth(r, tunnel)
	case ModeTokenProtected:
		return enforceTunnelToken(r, tunnel)
	default:
		return 0, "", true
	}
}

func enforceTunnelBasicAuth(r *http.Request, tunnel registry.Tunnel) (int, string, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Basic ") {
		return http.StatusUnauthorized, "missing basic auth", false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Basic "))
	if err != nil {
		return http.StatusUnauthorized, "invalid basic auth", false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return http.StatusUnauthorized, "invalid basic auth", false
	}
	if subtle.ConstantTimeCompare([]byte(parts[0]), []byte(tunnel.BasicAuthUsername)) != 1 ||
		subtle.ConstantTimeCompare([]byte(parts[1]), []byte(tunnel.BasicAuthPassword)) != 1 {
		return http.StatusUnauthorized, "invalid basic auth", false
	}
	return 0, "", true
}

func enforceTunnelToken(r *http.Request, tunnel registry.Tunnel) (int, string, bool) {
	provided := r.Header.Get("X-Bloop-Token")
	if provided == "" {
		provided = r.URL.Query().Get("token")
	}
	if tunnel.ProtectedToken == "" || subtle.ConstantTimeCompare([]byte(tunnel.ProtectedToken), []byte(provided)) != 1 {
		return http.StatusUnauthorized, "invalid token", false
	}
	return 0, "", true
}
