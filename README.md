# bloop-tunnel

Private-first HTTP tunnel system for exposing local services through `bloop.to` using a VPS relay and a local client.

## Status

Early implementation. See `specs/001-v1-http-tunnels/` for spec, plan, and tasks.

## MVP Goal

Get a working path for:

public hostname -> relay -> client -> local HTTP service

## Planned binaries

- `bloop-relay`
- `bloop-client`

## Runtime ingest (v1)

The production-shaped ingest path is now client-owned:
- the client enrolls with control-plane
- the client receives a scoped ingest token
- the client reports runtime truth directly to control-plane

Relay-side ingest support still exists for local/dev compatibility, but the main release path is client-side bearer ingest.

## Install the client
See `docs/CLIENT_INSTALL.md` for:
- Docker install/run
- native macOS/Linux/Windows usage
- config examples
- verification steps
- AI agent / automation hints

## Public release artifacts
This repo now includes:
- CI workflow: `.github/workflows/ci.yml`
- release workflow: `.github/workflows/release.yml`

Release workflow behavior:
- builds the public `bloop-relay` binary for Linux/macOS/Windows
- uploads binaries to the GitHub Release
- publishes the public container image to both GHCR and Docker Hub:
  - `ghcr.io/<owner>/bloop-relay:<tag>`
  - `docker.io/<dockerhub-user>/bloop-relay:<tag>`
- Docker publishing uses `docker/build-push-action` with explicit multi-registry tags for more reliable GHCR + Docker Hub publication
- release workflow also syncs the Docker Hub repository short description from the GitHub repository description

Versioning / release policy:
- tags use semver: `vMAJOR.MINOR.PATCH`
- current default automation starts at `v0.1.0`
- every successful CI run on `develop` automatically creates the next prerelease tag (`v0.1.1-rc.1`, etc.)
- every successful CI run on `main` automatically creates the next stable patch release
- recommended flow: feature branch -> PR to `develop` -> validate prerelease -> promote to `main`

## Local relay/client ingest integration

For local end-to-end dev proofing, this repo includes:
- `deploy/compose/dev-relay-ingest.yml`
- `deploy/examples/relay.local.yaml`
- `deploy/examples/client.relay-ingest.yaml`

That stack runs:
- `bloop-relay`
- `bloop-client`
- a tiny local echo server target

It is intended to pair with the control-plane repo's `deploy/compose/dev-full.yml` and `make dev-runtime-e2e` workflow.
