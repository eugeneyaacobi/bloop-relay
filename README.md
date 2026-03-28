# bloop-relay

Server-side relay for the bloop tunnel system.

This repository is the **server/infrastructure package**. It runs on a VPS or public host, accepts client connections, manages hostname registrations, enforces access policy, and proxies public traffic to connected clients.

## Status

Active server-side relay repository for the bloop tunnel system.

## What this repo ships

This repo's public/server install and release story is centered on one binary:

- `bloop-relay` — the public relay/server that accepts tunnel client connections and serves inbound traffic

## Responsibilities

`bloop-relay` is responsible for:
- accepting persistent client sessions
- authenticating client connections
- managing hostname leases and tunnel registrations
- enforcing public/basic-auth/token-based access rules
- proxying public HTTP traffic to the correct connected client
- optional server-side runtime snapshot reporting

## Install / operate the relay

See:
- `docs/VPS_DEPLOYMENT.md`

## CI and release artifacts

GitHub Actions handles verification and relay release packaging:

- `.github/workflows/ci.yml`
  - runs `go test ./...`
  - builds the relay binary
  - builds the relay image

- `.github/workflows/release.yml`
  - triggers on `v*` tags, published releases, or manual dispatch
  - builds versioned native archives for `bloop-relay`
  - uploads release assets to the GitHub Release
  - publishes the `bloop-relay` container image to GHCR and Docker Hub

### Expected GitHub release assets

Each release should include archives like:
- `bloop-relay-vX.Y.Z-linux-amd64.tar.gz`
- `bloop-relay-vX.Y.Z-linux-arm64.tar.gz`
- `bloop-relay-vX.Y.Z-darwin-amd64.tar.gz`
- `bloop-relay-vX.Y.Z-darwin-arm64.tar.gz`
- `bloop-relay-vX.Y.Z-windows-amd64.zip`

## Local build examples

Build the relay locally:

```bash
go build -o dist/bloop-relay ./cmd/bloop-relay
```

Build the relay container locally:

```bash
docker build -f deploy/docker/relay.Dockerfile -t bloop-relay:local .
```

## Relationship to bloop-tunnel

The end-user local client now belongs in the separate `bloop-tunnel` repository.

This repository is intentionally server-side only: relay runtime, server deployment assets, and relay release automation.
