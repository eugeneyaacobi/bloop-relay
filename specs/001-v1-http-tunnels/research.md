# Research Notes: bloop-tunnel v1 HTTP tunnel system

## Chosen Direction

- Use Go for both relay and client.
- Use WebSocket for the persistent relay-client session transport.
- Keep V1 HTTP/HTTPS only.
- Keep TLS termination at the external edge proxy.
- Keep relay tunnel registry in memory for V1.
- Keep operator config file-driven.

## Alternatives Considered

### frp-style broader protocol support in V1
Rejected because it expands scope too early. HTTP/HTTPS alone delivers immediate value with less protocol complexity.

### Raw TCP transport between client and relay
Rejected for V1 because WebSocket is simpler through NAT/firewalls and easier to debug in a Docker-first personal deployment.

### Built-in TLS management in relay
Rejected because external edge tools already solve certificates and wildcard routing better than a fresh internal implementation.

### Dashboard-first management
Rejected because file-based config and logs are enough for a private-first operator workflow.

### Python implementation
Rejected because Go is a stronger fit for a small, portable networking binary pair and smoother Docker deployment.
