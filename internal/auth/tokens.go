package auth

import (
	"os"

	"bloop-tunnel/internal/config"
)

func ResolveRelayToken(t config.RelayToken) string {
	if t.Token != "" {
		return t.Token
	}
	if t.TokenEnv != "" {
		return os.Getenv(t.TokenEnv)
	}
	return ""
}

func ResolveClientToken(raw, envName string) string {
	if raw != "" {
		return raw
	}
	if envName != "" {
		return os.Getenv(envName)
	}
	return ""
}
