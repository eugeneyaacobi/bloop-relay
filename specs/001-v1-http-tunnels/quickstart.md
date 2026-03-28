# Quickstart: Deploy bloop-tunnel on a VPS for `bloop.to`

This guide gets the V1 relay running on a new VPS and connects a local client from your laptop.

## Architecture

- **VPS** runs:
  - Caddy (TLS + wildcard edge)
  - bloop-relay container
- **Laptop** runs:
  - bloop-client container
  - local services on the host machine
- **Cloudflare** provides DNS for `bloop.to`

## Prerequisites

- A VPS running Linux with Docker installed
- Control of the `bloop.to` DNS zone in Cloudflare
- A Cloudflare API token with DNS edit permission for the zone
- One relay auth token you will share with your laptop client

## DNS Setup

### Option A: Wildcard DNS (recommended)

Create a wildcard DNS record in Cloudflare:

- Type: `A`
- Name: `*`
- Value: your VPS public IP

Optional but recommended:
- Type: `A`
- Name: `relay`
- Value: your VPS public IP

**Proxying**:
- Start with Cloudflare DNS-only (grey cloud) while debugging
- Move to proxied mode later if desired and compatible with your traffic patterns

## VPS Directory Layout

Suggested layout on the server:

```text
/opt/bloop-tunnel/
├── compose/
│   └── relay-compose.yml
├── config/
│   └── relay.yaml
└── caddy/
    └── Caddyfile
```

## Relay Config

Example `/opt/bloop-tunnel/config/relay.yaml`:

```yaml
domain: bloop.to
listen_addr: ":8080"
trusted_proxies:
  - 127.0.0.1/32
client_tokens:
  - name: laptop-main
    token_env: BLOOP_CLIENT_TOKEN
hostname_generation:
  mode: random
  length: 8
logging:
  level: info
  format: json
```

## Caddy Config

Example `/opt/bloop-tunnel/caddy/Caddyfile`:

```caddy
*.bloop.to {
    tls {
        dns cloudflare {env.CLOUDFLARE_API_TOKEN}
    }

    reverse_proxy relay:8080
}
```

## Docker Compose (VPS)

Example `/opt/bloop-tunnel/compose/relay-compose.yml`:

```yaml
services:
  relay:
    build:
      context: /opt/bloop-tunnel/repo
      dockerfile: deploy/docker/relay.Dockerfile
    command: ["--config", "/config/relay.yaml"]
    environment:
      BLOOP_CLIENT_TOKEN: "replace-with-long-random-token"
    volumes:
      - /opt/bloop-tunnel/config/relay.yaml:/config/relay.yaml:ro
    restart: unless-stopped

  caddy:
    image: caddy:2
    ports:
      - "80:80"
      - "443:443"
    environment:
      CLOUDFLARE_API_TOKEN: "replace-with-cloudflare-token"
    volumes:
      - /opt/bloop-tunnel/caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - relay
    restart: unless-stopped

volumes:
  caddy_data:
  caddy_config:
```

## Laptop Client Config

Example `client.yaml`:

```yaml
relay_url: wss://relay.bloop.to/connect
auth_token_env: BLOOP_CLIENT_TOKEN
reconnect:
  initial_delay_ms: 1000
  max_delay_ms: 30000
logging:
  level: info
  format: json
tunnels:
  - name: app
    hostname: app.bloop.to
    local_addr: host.docker.internal:3000
    access: public

  - name: admin
    hostname: admin.bloop.to
    local_addr: host.docker.internal:4000
    access: basic_auth
    basic_auth:
      username: gene
      password_env: BLOOP_ADMIN_PASSWORD
```

## Laptop Docker Run Example

```bash
docker run --rm \
  --add-host host.docker.internal:host-gateway \
  -e BLOOP_CLIENT_TOKEN='replace-with-long-random-token' \
  -e BLOOP_ADMIN_PASSWORD='replace-with-strong-password' \
  -v "$PWD/client.yaml:/config/client.yaml:ro" \
  bloop-client:latest \
  --config /config/client.yaml
```

## First Bring-Up Checklist

### On VPS

1. Clone or copy the repository to the server
2. Put relay config in `/opt/bloop-tunnel/config/relay.yaml`
3. Put Caddyfile in `/opt/bloop-tunnel/caddy/Caddyfile`
4. Set strong values for:
   - `BLOOP_CLIENT_TOKEN`
   - `CLOUDFLARE_API_TOKEN`
5. Start the stack:

```bash
docker compose -f /opt/bloop-tunnel/compose/relay-compose.yml up -d --build
```

### On laptop

1. Create `client.yaml`
2. Export the same `BLOOP_CLIENT_TOKEN`
3. Start your local web app
4. Run the client container

## Verification

### Confirm relay health

```bash
docker compose -f /opt/bloop-tunnel/compose/relay-compose.yml logs -f relay
```

You should see relay startup and client connection logs.

### Confirm client registration

Client logs should show successful registration.

### Confirm public tunnel

```bash
curl -i https://app.bloop.to/
```

### Confirm protected tunnel

```bash
curl -i -u gene:YOUR_PASSWORD https://admin.bloop.to/
```

## Debugging Notes

- If Caddy cannot issue certificates, verify Cloudflare token scope and DNS records.
- If the client cannot reach the local service, verify the host address works from inside Docker.
- If the relay shows no client connection, verify firewall rules and that `/connect` is reachable through the edge.
- If wildcard routing behaves strangely, temporarily use explicit hostnames and DNS-only mode in Cloudflare.

## Security Notes

- Use long random relay auth tokens.
- Do not commit relay/client secrets into git.
- Prefer protected tunnels for anything non-public.
- Start with Cloudflare DNS-only until behavior is verified.
- Treat the relay as internet-facing infrastructure and keep the VPS patched.
