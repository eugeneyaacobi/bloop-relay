package access

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"bloop-tunnel/internal/config"
)

func Enforce(r *http.Request, tunnel config.TunnelConfig) (int, string, bool) {
	switch ResolveMode(tunnel.Access) {
	case ModeBasicAuth:
		return enforceBasicAuth(r, tunnel)
	case ModeTokenProtected:
		return enforceToken(r, tunnel)
	default:
		return 0, "", true
	}
}

func enforceBasicAuth(r *http.Request, tunnel config.TunnelConfig) (int, string, bool) {
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
	password := tunnel.BasicAuth.Password
	if password == "" && tunnel.BasicAuth.PasswordEnv != "" {
		password = os.Getenv(tunnel.BasicAuth.PasswordEnv)
	}
	if subtle.ConstantTimeCompare([]byte(parts[0]), []byte(tunnel.BasicAuth.Username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(parts[1]), []byte(password)) != 1 {
		return http.StatusUnauthorized, "invalid basic auth", false
	}
	return 0, "", true
}

func enforceToken(r *http.Request, tunnel config.TunnelConfig) (int, string, bool) {
	expected := tunnel.Token
	if expected == "" && tunnel.TokenEnv != "" {
		expected = os.Getenv(tunnel.TokenEnv)
	}
	provided := r.Header.Get("X-Bloop-Token")
	if provided == "" {
		provided = r.URL.Query().Get("token")
	}
	if expected == "" || subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) != 1 {
		return http.StatusUnauthorized, "invalid token", false
	}
	return 0, "", true
}
