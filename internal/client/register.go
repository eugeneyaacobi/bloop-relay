package client

import (
	"os"

	"bloop-tunnel/internal/protocol"
)

func (s *Session) RegisterTunnels() error {
	regs := make([]protocol.TunnelRegistration, 0, len(s.Config.Tunnels))
	for _, t := range s.Config.Tunnels {
		password := t.BasicAuth.Password
		if password == "" && t.BasicAuth.PasswordEnv != "" {
			password = os.Getenv(t.BasicAuth.PasswordEnv)
		}
		protectedToken := t.Token
		if protectedToken == "" && t.TokenEnv != "" {
			protectedToken = os.Getenv(t.TokenEnv)
		}

		regs = append(regs, protocol.TunnelRegistration{
			Name:              t.Name,
			Hostname:          t.Hostname,
			LocalAddr:         t.LocalAddr,
			Access:            t.Access,
			BasicAuthUsername: t.BasicAuth.Username,
			BasicAuthPassword: password,
			ProtectedToken:    protectedToken,
		})
	}

	msg := protocol.Envelope{
		Type: protocol.TypeRegisterTunnels,
		Payload: protocol.RegisterTunnels{
			Tunnels: regs,
		},
	}
	if err := s.Transport.Conn.WriteJSON(msg); err != nil {
		return err
	}

	var reply protocol.Envelope
	if err := s.Transport.Conn.ReadJSON(&reply); err != nil {
		return err
	}

	s.registered = regs
	s.Logger.Info("registered tunnels", "count", len(regs))
	return nil
}
