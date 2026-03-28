# VPS Deployment Notes

This document summarizes the intended first production-style deployment for `bloop-relay`.

## Recommended Stack

- Ubuntu or Debian VPS
- Docker + Docker Compose
- Caddy for wildcard TLS and reverse proxy
- Cloudflare wildcard DNS

## Ports

- 80/tcp for ACME HTTP challenge / redirects
- 443/tcp for HTTPS public traffic
- relay container internal port 8080 only (no need to expose publicly if fronted by Caddy)

## Files to prepare

- relay config YAML
- Caddyfile
- docker compose file
- environment variables for tokens

## Suggested system checks

- `docker version`
- `docker compose version`
- `ufw status` or equivalent firewall status
- DNS resolves VPS IP for `relay.bloop.to` and wildcard hostnames

## Suggested first test sequence

1. bring up relay + caddy
2. tail relay logs
3. connect client from laptop
4. verify public hostname
5. verify protected hostname
6. restart relay and verify client recovery

## Future production improvements

- systemd wrappers or compose restart policy review
- persistent relay state if hostname reservation continuity becomes important
- admin API for health and active tunnel visibility
- frontend control plane
