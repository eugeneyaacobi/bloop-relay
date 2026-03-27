# bloop-tunnel Constitution

## Core Principles

### I. Personal Infrastructure First
bloop-tunnel is built first as a personal tool for exposing local development services through Gene-controlled infrastructure. Design choices MUST prefer operator clarity, easy recovery, and low surprise over generic SaaS flexibility.

### II. Small, Composable Surface Area
The system MUST begin with the smallest useful product: HTTP/HTTPS tunnel exposure through a public relay and local client. New protocol support, dashboards, and multi-tenant features are out of scope unless explicitly added in later specs.

### III. Secure by Default
All network-facing features MUST have explicit access semantics. Every tunnel MUST be either public or protected. Protected tunnels MUST support at least basic authentication or token-based access. Secrets MUST never be hard-coded in source control.

### IV. External Edge, Simple Core
TLS termination, wildcard certificates, and internet edge concerns SHOULD be handled by an external reverse proxy such as Caddy or Traefik in front of the relay. The application core SHOULD focus on registration, routing, forwarding, and policy rather than edge certificate orchestration.

### V. Operator-Friendly Configuration
Users MUST be able to define a hostname explicitly or let the system generate one when omitted. Configuration and runtime behavior MUST be understandable from files and logs without requiring a web UI.

### VI. Testable Client/Server Contracts
The relay server and local client MUST communicate through explicit, versioned contracts that can be integration-tested. Reconnection, registration, tunnel lookup, and request forwarding behavior MUST be testable without manual inspection.

### VII. Open-Source Readiness Without Premature Complexity
The initial build may optimize for private use, but code structure, documentation, and secret handling SHOULD keep eventual open-source publication realistic. Avoid private-only assumptions that would require a rewrite later.

## Scope Constraints

- V1 targets HTTP/HTTPS only.
- V1 supports mixed access modes: public and protected.
- V1 supports user-specified hostnames and generated hostnames.
- V1 is deployed with Docker containers.
- V1 supports a macOS laptop first but SHOULD keep client/server behavior portable across macOS, Linux, and Windows.
- Raw TCP, UDP, browser UI, billing, and multi-user SaaS features are out of scope for V1.

## Development Workflow

- Start with constitution, then specification, then plan, then tasks.
- Favor integration tests around client/server behavior before broad feature expansion.
- Keep artifacts explicit: spec, plan, tasks, config examples, deployment examples.
- Each feature MUST preserve a working local development path using Docker.

## Governance
This constitution governs all design and implementation decisions in this repository. Simplicity beats cleverness. Scope discipline beats feature sprawl. Any exception requires an explicit spec update explaining why the new complexity is worth it.

**Version**: 1.0.0 | **Ratified**: 2026-03-26 | **Last Amended**: 2026-03-26
