# Implementation Plan: bloop-tunnel v1 HTTP tunnel system

**Branch**: `[001-v1-http-tunnels]` | **Date**: 2026-03-26 | **Spec**: `/root/.openclaw/workspace/bloop-tunnel/specs/001-v1-http-tunnels/spec.md`
**Input**: Feature specification from `/specs/001-v1-http-tunnels/spec.md`

## Summary

Build a small, private-first HTTP/HTTPS tunneling system for `bloop.to` using a Docker-deployed public relay and a Dockerized local client. The external edge handles TLS and wildcard hostname routing, while the application core focuses on client authentication, tunnel registration, hostname assignment, access-policy enforcement, and HTTP request forwarding from public hostnames to local laptop services.

The implementation will use Go for both relay and client, with a shared protocol package and integration-testable contracts. V1 deliberately supports HTTP/HTTPS only, mixed public/protected tunnel modes, user-specified or generated hostnames, automatic reconnection, and file-based configuration.

## Technical Context

**Language/Version**: Go 1.24+  
**Primary Dependencies**: Go standard library, Gorilla WebSocket (or nhooyr/websocket equivalent), Cobra for CLI, Viper or koanf for configuration loading, Testcontainers-Go for integration testing  
**Storage**: In-memory state for V1 runtime tunnel registry; file-based config; optional lightweight persistence deferred  
**Testing**: Go test, table-driven unit tests, contract tests, Docker-backed integration tests with Testcontainers  
**Target Platform**: Linux VPS for relay; Dockerized client intended to work on macOS, Linux, and Windows hosts  
**Project Type**: Networked CLI + background service pair (relay server + local client)  
**Performance Goals**: tunnel registration completes within 10 seconds; reconnection recovery within 30 seconds; reliable forwarding for normal local development traffic  
**Constraints**: HTTP/HTTPS only in V1; no required web UI; TLS terminated externally; secrets must not be committed; config must work in Docker deployments  
**Scale/Scope**: single-operator private deployment first; one relay; multiple tunnels per client; future OSS publication without major rewrite

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Gate Review

- **Personal Infrastructure First**: Pass. Architecture is operator-centric and private-first.
- **Small, Composable Surface Area**: Pass. V1 remains HTTP/HTTPS-only with no SaaS control plane or UI.
- **Secure by Default**: Pass. Access policy is explicit per tunnel and secret handling is externalized.
- **External Edge, Simple Core**: Pass. TLS and wildcard certificate concerns stay in Caddy/Traefik.
- **Operator-Friendly Configuration**: Pass. File-driven relay/client config is first-class.
- **Testable Client/Server Contracts**: Pass. Shared protocol and integration tests are part of the design.
- **Open-Source Readiness Without Premature Complexity**: Pass. Private-first scope is preserved without hard-coding private-only assumptions.

No constitution violations currently require justification.

## Project Structure

### Documentation (this feature)

```text
specs/001-v1-http-tunnels/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── relay-api.md
│   └── client-session-protocol.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── bloop-relay/
└── bloop-client/

internal/
├── auth/
├── client/
├── config/
├── generator/
├── logging/
├── protocol/
├── relay/
│   ├── access/
│   ├── registry/
│   ├── routing/
│   └── session/
└── transport/

pkg/
└── version/

deploy/
├── docker/
│   ├── relay.Dockerfile
│   └── client.Dockerfile
├── compose/
│   ├── relay-compose.yml
│   └── example-client-compose.yml
└── examples/
    ├── Caddyfile
    ├── traefik-dynamic.yml
    ├── relay.example.yaml
    └── client.example.yaml

test/
├── contract/
├── integration/
└── fixtures/
```

**Structure Decision**: Use a single Go module with two binaries (`bloop-relay` and `bloop-client`) and shared internal packages. Keep deployment examples and documentation in-repo so private use and future open-source publication share the same artifact trail.

## Architecture Overview

### Runtime Components

1. **External Edge Proxy**
   - Runs on the VPS in front of the relay.
   - Terminates TLS for `*.bloop.to`.
   - Forwards HTTP traffic to the relay with the original Host header preserved.

2. **Relay Server**
   - Public-facing control and routing service.
   - Authenticates client sessions using pre-shared tokens.
   - Maintains active tunnel registry in memory.
   - Resolves inbound hostname → tunnel mapping.
   - Enforces access policies.
   - Forwards request payloads over the persistent client session.

3. **Local Client**
   - Runs in Docker on the laptop.
   - Reads tunnel definitions from config.
   - Opens a persistent outbound control connection to the relay.
   - Registers multiple tunnel definitions.
   - Proxies requests received from relay to local HTTP targets.
   - Reconnects and re-registers automatically.

### Request Flow

1. Operator starts relay on VPS behind Caddy/Traefik.
2. Operator starts client on laptop with auth token and tunnel config.
3. Client connects to relay over a persistent session.
4. Client registers one or more HTTP tunnels.
5. Relay accepts public requests from edge proxy and selects the matching tunnel by Host header.
6. Relay enforces public/basic-auth/token access policy.
7. Relay forwards the request to the owning client over the session transport.
8. Client proxies the request to the configured local address.
9. Client streams response metadata/body back to relay.
10. Relay returns the response to the public caller.

## Protocol Design

### Session Transport

Use a long-lived WebSocket connection between client and relay for V1.

Reasoning:
- outbound-friendly through NAT/firewalls
- simpler than custom TCP multiplexing for the first version
- good enough for HTTP request/response forwarding
- debuggable and portable

### Session Message Types

Initial protocol shape:
- `client_hello`
- `server_hello`
- `register_tunnels`
- `register_result`
- `unregister_tunnels`
- `request_begin`
- `request_body_chunk`
- `request_end`
- `response_begin`
- `response_body_chunk`
- `response_end`
- `error`
- `ping`
- `pong`

### Protocol Rules

- Each forwarded public request gets a unique request ID.
- One client session may own multiple active tunnels.
- Hostname registration is atomic per registration attempt.
- Unknown or conflicting hostnames fail registration cleanly.
- Relay is authoritative for tunnel hostname leases.
- Client re-registration after reconnect is idempotent.

## Configuration Model

### Relay Config

Fields:
- listen address
- trusted domain suffix (e.g. `bloop.to`)
- allowed generated hostname pattern settings
- client auth token list or token source
- edge/proxy trust settings
- log format / level

Example:

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

### Client Config

Fields:
- relay URL
- auth token / token env ref
- reconnect policy
- list of tunnel definitions

Tunnel fields:
- name
- hostname (optional)
- local address
- access mode (`public`, `basic_auth`, `token_protected`)
- credentials/token refs

Example:

```yaml
relay_url: wss://relay.bloop.to/connect
auth_token_env: BLOOP_CLIENT_TOKEN
reconnect:
  initial_delay_ms: 1000
  max_delay_ms: 30000

tunnels:
  - name: app
    hostname: app.bloop.to
    local_addr: host.docker.internal:3000
    access: public

  - name: webhook
    local_addr: host.docker.internal:8787
    access: basic_auth
    basic_auth:
      username: gene
      password_env: BLOOP_WEBHOOK_PASSWORD

  - name: admin
    hostname: admin.bloop.to
    local_addr: host.docker.internal:9000
    access: token_protected
    token_env: BLOOP_ADMIN_TOKEN
```

## Data Model

### Tunnel
- ID
- name
- hostname
- access mode
- local address
- client session ID
- registration timestamp
- status

### Client Session
- session ID
- client name / token identity
- connected at
- last heartbeat
- registered tunnel IDs
- connection state

### Hostname Lease
- hostname
- owning session ID
- tunnel ID
- lease status
- created at

## Security Design

- Relay accepts tunnel registrations only from authenticated clients.
- Every tunnel declares an access mode explicitly.
- Protected tunnels enforce auth at relay before forwarding.
- Secrets are injected through environment variables or external config files.
- Relay trusts the external reverse proxy only when configured explicitly.
- Unknown public hostnames return a safe error response.
- Logs must avoid printing raw credentials or tokens.

## Deployment Design

### VPS Stack

Recommended stack:
- Caddy or Traefik
- bloop-relay container
- optional Docker Compose

Responsibilities:
- wildcard TLS for `*.bloop.to`
- forwarding all matching hostnames to relay
- preserving `Host` and forwarding headers

### Laptop Stack

Recommended stack:
- bloop-client container
- host service(s) running outside or alongside Docker

Cross-platform note:
- default docs should use `host.docker.internal` where available
- document Linux alternatives for reaching host services from Docker

## Testing Strategy

### Unit Tests
- hostname generation behavior
- access policy evaluation
- config parsing and validation
- hostname conflict detection
- reconnect backoff policy

### Contract Tests
- session message encode/decode
- registration success/failure cases
- request/response message lifecycle

### Integration Tests
- relay + client registration round-trip
- public tunnel forwarding to local test HTTP server
- basic-auth enforcement
- token-protected tunnel enforcement
- reconnect + re-register behavior
- hostname conflict rejection

### Manual Verification
- Docker Compose relay deployment on VPS-like environment
- local client on macOS host reaching host HTTP service

## Implementation Phases

### Phase 0 - Artifact completion
- Write research.md capturing selected architectural tradeoffs and alternatives rejected.
- Write contracts for relay HTTP ingress assumptions and client session protocol.
- Write data-model.md for relay/client entities.
- Write quickstart.md with Docker + edge proxy setup.

### Phase 1 - Core shared foundations
- Initialize Go module.
- Create shared config loader.
- Create protocol package and message schema.
- Create logging/version packages.
- Add initial contract tests.

### Phase 2 - Relay registration and routing core
- Implement relay session manager.
- Implement client authentication.
- Implement tunnel registry and hostname lease checks.
- Implement edge HTTP ingress path and hostname resolution.

### Phase 3 - Client session + forwarding
- Implement persistent client session connection.
- Implement tunnel registration flow.
- Implement request forwarding from relay to client and local HTTP target.
- Implement response streaming back to relay.

### Phase 4 - Access controls + resilience
- Implement `public`, `basic_auth`, and `token_protected` modes.
- Implement reconnect backoff.
- Implement idempotent re-registration after reconnect.
- Add integration tests for auth and recovery.

### Phase 5 - Deployment + operator polish
- Add Dockerfiles and Compose examples.
- Add Caddy/Traefik examples.
- Add example relay/client configs.
- Finalize quickstart and operational documentation.

## Open Questions

These are not blockers for the plan but should be resolved during implementation artifacts:
- Should generated hostnames be random-only or optionally name-derived with conflict fallback?
- Should relay return 404, 502, or branded error pages for unknown or unhealthy tunnel targets?
- Should V1 support streaming request/response bodies with chunked forwarding from day one, or constrain initial body size for MVP simplicity?

## Complexity Tracking

No constitution violations currently require justification.
