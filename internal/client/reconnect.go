package client

import (
	"context"
	"time"
)

func (s *Session) RunWithReconnect(ctx context.Context) error {
	delay := time.Duration(s.Config.Reconnect.InitialDelayMs) * time.Millisecond
	maxDelay := time.Duration(s.Config.Reconnect.MaxDelayMs) * time.Millisecond
	if delay <= 0 {
		delay = time.Second
	}
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}

	for {
		err := s.Run()
		if err == nil {
			return nil
		}
		s.Logger.Warn("session disconnected, attempting reconnect", "error", err, "delay", delay.String())

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		next, err := Connect(ctx, s.Config, s.Logger)
		if err != nil {
			s.Logger.Warn("reconnect failed", "error", err)
			if delay < maxDelay {
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
			}
			continue
		}

		s.Transport = next.Transport
		if err := s.RegisterTunnels(); err != nil {
			s.Logger.Warn("re-register tunnels failed", "error", err)
			_ = s.Transport.Close()
			if delay < maxDelay {
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
			}
			continue
		}

		delay = time.Duration(s.Config.Reconnect.InitialDelayMs) * time.Millisecond
		if delay <= 0 {
			delay = time.Second
		}
		s.Logger.Info("reconnected and re-registered tunnels")
	}
}
