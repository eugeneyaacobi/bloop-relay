# Feature Specification: bloop-tunnel v1 HTTP tunnel system

**Feature Branch**: `[001-v1-http-tunnels]`  
**Created**: 2026-03-26  
**Status**: Draft  
**Input**: User description: "Create a small tool for bloop.to that runs in Docker, exposes local laptop ports through a VPS relay, learns from LocalToNet and similar tools, starts private but can later become open source, uses wildcard Cloudflare DNS when available, lets users specify hostnames or have them generated, supports mixed public/protected access, and targets HTTP/HTTPS first."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Expose a local web app on bloop.to (Priority: P1)

As the operator, I want to run a local client on my laptop, register a local HTTP service, and receive a working public HTTPS hostname on `bloop.to` so I can access or share my local app through my own infrastructure.

**Why this priority**: This is the core product value. If this works reliably, the tool is already useful.

**Independent Test**: Can be fully tested by running the relay and client with Docker, exposing a local development server, and successfully loading the public hostname through the VPS.

**Acceptance Scenarios**:

1. **Given** a running relay, a configured wildcard DNS setup, and a local web service, **When** the client registers a tunnel with a specified hostname, **Then** requests to that hostname are forwarded to the local service and return successful HTTP responses.
2. **Given** a running relay and a local web service, **When** the client registers a tunnel without a hostname, **Then** the system generates a hostname under the configured domain and makes the local service reachable there.

---

### User Story 2 - Protect a tunnel from unwanted access (Priority: P2)

As the operator, I want a tunnel to be either public or protected so that I can selectively expose local services without making everything openly reachable.

**Why this priority**: Security posture matters immediately once a local service is internet-facing.

**Independent Test**: Can be fully tested by creating one public tunnel and one protected tunnel, then verifying that the public tunnel succeeds unauthenticated while the protected one rejects unauthorized requests and accepts valid credentials.

**Acceptance Scenarios**:

1. **Given** a protected tunnel using basic authentication, **When** a request is sent without credentials, **Then** the relay denies access.
2. **Given** a protected tunnel using token-based protection, **When** a request includes the valid configured token, **Then** the request is forwarded successfully.
3. **Given** a public tunnel, **When** a request is sent without credentials, **Then** the request is forwarded successfully.

---

### User Story 3 - Recover from disconnections without manual babysitting (Priority: P3)

As the operator, I want the client to reconnect automatically and restore its tunnel registrations after network interruptions so the tool remains practical on a laptop.

**Why this priority**: Laptop network changes and sleep/wake cycles are normal. Manual re-registration would make the tool annoying fast.

**Independent Test**: Can be fully tested by interrupting the client-to-relay connection, restoring connectivity, and verifying that tunnel registrations return automatically without manual reconfiguration.

**Acceptance Scenarios**:

1. **Given** an active registered tunnel, **When** the client connection drops temporarily and reconnects, **Then** the tunnel becomes reachable again without manual intervention.
2. **Given** the relay restarts while the client remains configured, **When** the client reconnects, **Then** the relay restores the client’s active tunnel registrations.

---

### Edge Cases

- What happens when a requested hostname is already claimed by another active tunnel?
- What happens when the local target port is unavailable or returns connection errors?
- How does the system behave when Cloudflare wildcard DNS exists but the requested hostname does not yet route correctly at the edge?
- What happens when a generated hostname collides with an existing active or reserved hostname?
- How does the relay handle oversized request bodies or long-lived streaming responses?
- What happens when the laptop sleeps and wakes repeatedly during an active session?
- How does the system behave when the reverse proxy forwards traffic for a hostname that is unknown to the relay?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a relay server that accepts persistent client connections from operator-controlled client agents.
- **FR-002**: The system MUST provide a local client that runs in Docker and forwards HTTP requests from the relay to configured local HTTP targets.
- **FR-003**: The system MUST allow a tunnel to be configured with a user-specified hostname under the configured parent domain.
- **FR-004**: The system MUST generate a hostname under the configured parent domain when the operator does not specify one.
- **FR-005**: The system MUST reject tunnel registration when a requested hostname is already in use by another active tunnel or reserved by policy.
- **FR-006**: The system MUST support HTTP/HTTPS application traffic in V1.
- **FR-007**: The system MUST support tunnel access modes of `public`, `basic_auth`, and `token_protected`.
- **FR-008**: The system MUST enforce the configured access mode before forwarding requests to the local target.
- **FR-009**: The system MUST allow one client instance to register multiple tunnels.
- **FR-010**: The system MUST reconnect clients automatically after transient connection loss.
- **FR-011**: The system MUST restore tunnel registrations automatically after client reconnection.
- **FR-012**: The system MUST provide operator-readable structured logs for tunnel registration, connection state changes, access denials, and forwarding failures.
- **FR-013**: The system MUST support deployment behind an external reverse proxy that performs TLS termination for the public domain.
- **FR-014**: The system MUST make hostname generation deterministic enough to avoid accidental duplicates within a relay instance.
- **FR-015**: The system MUST expose a configuration model that works in file-based Docker deployments without requiring a web UI.
- **FR-016**: The system MUST keep secrets configurable through environment variables or non-committed config files.
- **FR-017**: The system SHOULD preserve portability across macOS, Linux, and Windows for the client runtime.
- **FR-018**: The system MUST document a reference deployment pattern using a VPS, Docker, and an external reverse proxy.

### Key Entities *(include if feature involves data)*

- **Relay**: The public server component that authenticates clients, tracks tunnel registrations, enforces access rules, and routes inbound requests to the correct client session.
- **Client Session**: A persistent authenticated connection between a client instance and the relay, capable of owning multiple tunnel registrations.
- **Tunnel**: A mapping from a public hostname and access policy to a local HTTP target exposed by a specific client session.
- **Access Policy**: The protection mode attached to a tunnel, including public access, basic authentication, or token-based access.
- **Hostname Lease**: The relay-side record that marks a hostname as assigned or unavailable for conflicting registrations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator can deploy the relay and client with Docker, expose a local web service, and successfully serve it through a `bloop.to` hostname in under 30 minutes using the reference docs.
- **SC-002**: A newly created specified-hostname tunnel becomes reachable through the public relay within 10 seconds of successful client registration under normal network conditions.
- **SC-003**: A generated-hostname tunnel becomes reachable without manual DNS edits when wildcard DNS is already configured for the parent domain.
- **SC-004**: Protected tunnels reject unauthorized requests and accept valid credentials in 100% of automated acceptance tests.
- **SC-005**: After a transient client disconnect, an existing tunnel returns to service automatically within 30 seconds of connectivity restoration in automated reconnection tests.
- **SC-006**: The reference implementation runs the client and relay successfully in Docker on a development VPS + laptop setup without requiring a web dashboard.

## Assumptions

- Cloudflare-managed wildcard DNS will be configured by the operator for the chosen parent domain before production use.
- TLS termination will be handled by an external reverse proxy such as Caddy or Traefik rather than by the application itself in V1.
- V1 is intentionally limited to HTTP/HTTPS traffic; raw TCP and UDP are deferred.
- The initial operator is a single trusted owner rather than a multi-tenant customer base.
- The local client will usually forward to services running on the host machine reachable from the Docker container.
- A new VPS will be provisioned specifically to host the public relay stack.
- The initial release prioritizes private usability while keeping structure and docs suitable for future open-source publication.
