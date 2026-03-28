# bloop-relay

`bloop-relay` is the server-side relay for the bloop tunnel system.

This repository is the **operator-facing relay repository**: the code, container image, example configs, and deployment docs required to run the public ingress layer that accepts tunnel client connections and forwards traffic to connected agents.

## What this repo is for

Run this repository when you need to:
- host the public relay on a VPS, VM, or other internet-reachable server
- terminate inbound HTTP traffic for tunneled hostnames
- authenticate tunnel clients and register their sessions
- enforce access policy for exposed routes
- package and publish relay binaries and container images for operators

If you are looking for the end-user/local client runtime, use the separate `bloop-tunnel` repository. This repo is for the relay service and its deployment assets.

## What the relay does

`bloop-relay` is responsible for:
- accepting persistent client sessions
- authenticating client connections
- managing hostname leases and tunnel registrations
- enforcing public, basic-auth, and token-based access rules
- proxying public HTTP traffic to the correct connected client
- exposing the server runtime that operators deploy and monitor

## Release artifacts

Production releases are built around one operator-facing runtime:

- `bloop-relay` — the relay/server binary

The release workflow publishes:
- versioned native archives for supported OS/architecture pairs
- a multi-arch container image for Linux operators
- bundled operator docs and a starter relay config in each native archive

### Expected GitHub release assets

Each tagged release should include archives named like:
- `bloop-relay-vX.Y.Z-linux-amd64.tar.gz`
- `bloop-relay-vX.Y.Z-linux-arm64.tar.gz`
- `bloop-relay-vX.Y.Z-darwin-amd64.tar.gz`
- `bloop-relay-vX.Y.Z-darwin-arm64.tar.gz`
- `bloop-relay-vX.Y.Z-windows-amd64.zip`

Each native archive contains:
- the `bloop-relay` binary
- `README.md`
- `VPS_DEPLOYMENT.md`
- `deploy/examples/relay.example.yaml`

Container releases are published as:
- `ghcr.io/<owner>/bloop-relay:<tag>` and `:latest`
- `docker.io/<owner>/bloop-relay:<tag>` and `:latest`

## Deployment expectations

A production relay deployment should assume:
- a public hostname and working DNS
- TLS termination handled by a front proxy such as Caddy or Traefik
- only the relay service exposed behind that proxy
- relay configuration supplied via a checked-in config file or mounted secret-backed config
- client tokens managed as operator secrets, not hard-coded into images
- basic host-level controls in place: firewall rules, restart policy, log collection, and backups for any surrounding infrastructure

Start with:
- `docs/VPS_DEPLOYMENT.md`

## Repository layout for operators

Key paths:
- `deploy/docker/relay.Dockerfile` — container build for the relay
- `deploy/compose/relay-compose.yml` — compose example for running the relay
- `deploy/examples/relay.example.yaml` — starter relay configuration
- `docs/VPS_DEPLOYMENT.md` — deployment and operations guidance
- `docs/CLIENT_INSTALL.md` — pointer for client-side installs and test-only local client usage from this repo

## CI and release automation

GitHub Actions handles verification and packaging:

- `.github/workflows/ci.yml`
  - runs `go test ./...`
  - builds cross-platform relay binaries
  - verifies the relay container image builds

- `.github/workflows/release.yml`
  - triggers on `v*` tags, published releases, or manual dispatch
  - builds versioned native archives for `bloop-relay`
  - uploads release assets to the GitHub Release
  - publishes the `bloop-relay` container image to GHCR and Docker Hub

## Local build examples

Build the relay locally:

```bash
go build -o dist/bloop-relay ./cmd/bloop-relay
```

Build the relay container locally:

```bash
docker build -f deploy/docker/relay.Dockerfile -t bloop-relay:local .
```

## Summary

`bloop-relay` is the public relay service repo: server runtime, deployment assets, and operator release artifacts. It is not the primary home for the end-user client install story.
