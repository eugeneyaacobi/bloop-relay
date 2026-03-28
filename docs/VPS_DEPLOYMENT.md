# bloop-relay VPS deployment guide

This guide covers production-style deployment of `bloop-relay` as the internet-facing relay service for the bloop tunnel system.

## Deployment model

Recommended baseline:
- one VPS or VM with a stable public IP
- Docker Engine and Docker Compose
- a front proxy such as Caddy or Traefik handling TLS and public host routing
- DNS pointed at the server for the relay hostname and any public wildcard or managed hostnames

`bloop-relay` itself listens on an internal service port and is typically placed behind the front proxy.

## What operators should prepare

Before deployment, have:
- a server hostname or IP reachable from the public internet
- DNS records for the relay endpoint and any hostname patterns you plan to serve
- a relay config file based on `deploy/examples/relay.example.yaml`
- client authentication tokens or token environment variables
- a decision for where TLS terminates and how certificates are renewed
- firewall rules allowing only the ports you intend to expose

## Reference assets in this repo

Use these files as the operator starting point:
- `deploy/docker/relay.Dockerfile`
- `deploy/compose/relay-compose.yml`
- `deploy/examples/relay.example.yaml`
- `deploy/examples/Caddyfile`
- `deploy/examples/traefik-dynamic.yml`

## Network expectations

Typical port layout:
- `80/tcp` for ACME HTTP challenge or HTTP-to-HTTPS redirects
- `443/tcp` for public HTTPS traffic
- `8080/tcp` for the relay service behind the proxy

If the proxy and relay are on the same host, keep `8080` limited to local or private-network access where possible.

## Example relay configuration

The provided example config shows the core operator inputs:
- relay listen address
- trusted proxies
- client token sources
- hostname generation settings
- structured logging

Start from:

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

For production use:
- replace the example domain with your real domain
- scope trusted proxies to the actual reverse proxy addresses in front of the relay
- source tokens from environment variables or your secret manager
- prefer JSON logs for aggregation

## Container deployment flow

Build or pull the relay image:

```bash
docker pull ghcr.io/<owner>/bloop-relay:latest
```

Or build locally from this repo:

```bash
docker build -f deploy/docker/relay.Dockerfile -t bloop-relay:local .
```

The included compose example mounts a relay config and exposes port `8080`:

```yaml
services:
  relay:
    build:
      context: ../..
      dockerfile: deploy/docker/relay.Dockerfile
    command: ["--config", "/config/relay.yaml"]
    ports:
      - "8080:8080"
    volumes:
      - ../examples/relay.example.yaml:/config/relay.yaml:ro
    restart: unless-stopped
```

For production, update that example to:
- use a pinned release image instead of a local build
- mount your real config file
- inject required token environment variables
- place the service on a network shared with your reverse proxy
- add your preferred logging, monitoring, and restart settings

## Reverse proxy expectations

The relay should usually sit behind Caddy or Traefik so operators can:
- terminate TLS cleanly
- handle wildcard certificates
- keep public port management out of the relay process
- standardize headers and request logging

When running behind a proxy:
- ensure forwarded headers are correct
- set `trusted_proxies` to only the proxy addresses you control
- verify hostname routing reaches the relay unchanged

## Verification checklist

Before calling the deployment ready, verify:
- the relay container or process starts cleanly
- the config file loads without errors
- the proxy can reach the relay on its internal port
- DNS resolves to the expected server
- TLS certificates are issued and renew successfully
- a test client can authenticate and connect
- a public hostname routes to the connected client
- protected routes enforce the intended access policy
- restart behavior is clean after a relay or host reboot

## Operational baseline

Minimum production posture:
- keep the host patched
- restrict inbound ports with a firewall or cloud security group
- run the relay with automatic restart behavior
- ship logs somewhere durable or aggregate them centrally
- monitor process health and basic request/connectivity signals
- document token rotation and server replacement steps

## Release artifact expectations

Tagged releases should give operators everything needed to start a deployment decision:
- native `bloop-relay` archives for supported platforms
- container images for Linux deployment
- bundled deployment docs
- a starter relay config example

That means this repo’s release story is server-first: operators download or pull the relay runtime, apply their config, and deploy it behind a public proxy.
