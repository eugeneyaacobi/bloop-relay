# Tasks: bloop-tunnel v1 HTTP tunnel system

**Input**: Design documents from `/specs/001-v1-http-tunnels/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Integration and contract tests are included because reliability, auth enforcement, and reconnection are core requirements.

**Organization**: Tasks are grouped by setup, foundational work, and user stories so each story can be implemented and tested independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (`US1`, `US2`, `US3`)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the repository structure, toolchain, and spec artifacts used by all later work.

- [ ] T001 Create Go module and baseline repository structure in `go.mod`, `cmd/`, `internal/`, `pkg/`, `deploy/`, and `test/`
- [ ] T002 [P] Add `.gitignore` entries for build outputs, local configs, secrets, and `.env` files in `/root/.openclaw/workspace/bloop-tunnel/.gitignore`
- [ ] T003 [P] Add Makefile or task runner commands for build, test, lint, and integration workflows in `/root/.openclaw/workspace/bloop-tunnel/Makefile`
- [ ] T004 [P] Add initial README project overview and development prerequisites in `/root/.openclaw/workspace/bloop-tunnel/README.md`
- [ ] T005 [P] Create contracts documentation files in `/root/.openclaw/workspace/bloop-tunnel/specs/001-v1-http-tunnels/contracts/relay-api.md` and `/root/.openclaw/workspace/bloop-tunnel/specs/001-v1-http-tunnels/contracts/client-session-protocol.md`
- [ ] T006 [P] Create quickstart deployment guide in `/root/.openclaw/workspace/bloop-tunnel/specs/001-v1-http-tunnels/quickstart.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before any user story implementation.

**⚠️ CRITICAL**: No user story work should begin until this phase is complete.

- [ ] T007 Implement shared version package in `/root/.openclaw/workspace/bloop-tunnel/pkg/version/version.go`
- [ ] T008 [P] Implement structured logging helpers in `/root/.openclaw/workspace/bloop-tunnel/internal/logging/logging.go`
- [ ] T009 [P] Implement relay config model and loader in `/root/.openclaw/workspace/bloop-tunnel/internal/config/relay.go`
- [ ] T010 [P] Implement client config model and loader in `/root/.openclaw/workspace/bloop-tunnel/internal/config/client.go`
- [ ] T011 Implement shared protocol message types in `/root/.openclaw/workspace/bloop-tunnel/internal/protocol/messages.go`
- [ ] T012 [P] Implement protocol serialization and validation helpers in `/root/.openclaw/workspace/bloop-tunnel/internal/protocol/codec.go`
- [ ] T013 [P] Implement auth token validation helpers in `/root/.openclaw/workspace/bloop-tunnel/internal/auth/tokens.go`
- [ ] T014 [P] Implement hostname generation helpers in `/root/.openclaw/workspace/bloop-tunnel/internal/generator/hostnames.go`
- [ ] T015 [P] Implement transport-level WebSocket session wrapper in `/root/.openclaw/workspace/bloop-tunnel/internal/transport/session.go`
- [ ] T016 [P] Add unit tests for config loading, auth helpers, and hostname generation in `/root/.openclaw/workspace/bloop-tunnel/internal/config/config_test.go`, `/root/.openclaw/workspace/bloop-tunnel/internal/auth/tokens_test.go`, and `/root/.openclaw/workspace/bloop-tunnel/internal/generator/hostnames_test.go`
- [ ] T017 [P] Add protocol contract tests for message encoding/decoding in `/root/.openclaw/workspace/bloop-tunnel/test/contract/protocol_contract_test.go`

**Checkpoint**: Shared foundations complete; user story implementation can begin.

---

## Phase 3: User Story 1 - Expose a local web app on bloop.to (Priority: P1) 🎯 MVP

**Goal**: Let a client register HTTP tunnels and serve a local web app through a public `bloop.to` hostname.

**Independent Test**: Start relay + client + local test HTTP service, then reach the service through the public hostname path handled by the relay.

### Tests for User Story 1

- [ ] T018 [P] [US1] Add relay/client integration test for successful tunnel registration in `/root/.openclaw/workspace/bloop-tunnel/test/integration/registration_test.go`
- [ ] T019 [P] [US1] Add integration test for specified hostname forwarding in `/root/.openclaw/workspace/bloop-tunnel/test/integration/http_forwarding_specified_hostname_test.go`
- [ ] T020 [P] [US1] Add integration test for generated hostname assignment in `/root/.openclaw/workspace/bloop-tunnel/test/integration/http_forwarding_generated_hostname_test.go`

### Implementation for User Story 1

- [ ] T021 [P] [US1] Implement relay session manager in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/session/manager.go`
- [ ] T022 [P] [US1] Implement tunnel registry and hostname lease tracking in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/registry/registry.go`
- [ ] T023 [P] [US1] Implement hostname routing lookup in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/routing/router.go`
- [ ] T024 [US1] Implement relay WebSocket connection handler and client hello/auth flow in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/session/handler.go`
- [ ] T025 [US1] Implement tunnel registration and conflict handling in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/registry/register.go`
- [ ] T026 [US1] Implement relay HTTP ingress handler that resolves hostname to tunnel in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/http_handler.go`
- [ ] T027 [P] [US1] Implement client session runtime and connection bootstrap in `/root/.openclaw/workspace/bloop-tunnel/internal/client/session.go`
- [ ] T028 [P] [US1] Implement client tunnel registration logic in `/root/.openclaw/workspace/bloop-tunnel/internal/client/register.go`
- [ ] T029 [US1] Implement request forwarding from relay to client and local HTTP target in `/root/.openclaw/workspace/bloop-tunnel/internal/client/forward_http.go`
- [ ] T030 [US1] Implement response streaming from client back to relay in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/response_proxy.go`
- [ ] T031 [US1] Create relay CLI entrypoint in `/root/.openclaw/workspace/bloop-tunnel/cmd/bloop-relay/main.go`
- [ ] T032 [US1] Create client CLI entrypoint in `/root/.openclaw/workspace/bloop-tunnel/cmd/bloop-client/main.go`
- [ ] T033 [US1] Add example relay and client configs in `/root/.openclaw/workspace/bloop-tunnel/deploy/examples/relay.example.yaml` and `/root/.openclaw/workspace/bloop-tunnel/deploy/examples/client.example.yaml`

**Checkpoint**: User Story 1 should now provide a working MVP: public hostname → relay → client → local HTTP service.

---

## Phase 4: User Story 2 - Protect a tunnel from unwanted access (Priority: P2)

**Goal**: Support explicit public and protected tunnel modes with relay-enforced access control.

**Independent Test**: Create one public tunnel and one protected tunnel, verify protected traffic is denied without credentials and accepted with correct credentials.

### Tests for User Story 2

- [ ] T034 [P] [US2] Add integration test for public tunnel access in `/root/.openclaw/workspace/bloop-tunnel/test/integration/public_access_test.go`
- [ ] T035 [P] [US2] Add integration test for basic auth enforcement in `/root/.openclaw/workspace/bloop-tunnel/test/integration/basic_auth_test.go`
- [ ] T036 [P] [US2] Add integration test for token-protected tunnel enforcement in `/root/.openclaw/workspace/bloop-tunnel/test/integration/token_access_test.go`

### Implementation for User Story 2

- [ ] T037 [P] [US2] Implement access policy model and evaluation helpers in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/access/policy.go`
- [ ] T038 [US2] Implement public/basic-auth/token-protected enforcement in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/access/enforce.go`
- [ ] T039 [US2] Wire access policy enforcement into relay ingress path in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/http_handler.go`
- [ ] T040 [US2] Add protected tunnel config validation in `/root/.openclaw/workspace/bloop-tunnel/internal/config/client.go`
- [ ] T041 [US2] Add redaction-safe logging for protected tunnel registration and auth failures in `/root/.openclaw/workspace/bloop-tunnel/internal/logging/logging.go`

**Checkpoint**: User Stories 1 and 2 should both work independently, with relay-enforced protection modes.

---

## Phase 5: User Story 3 - Recover from disconnections without manual babysitting (Priority: P3)

**Goal**: Automatically recover from client disconnects and restore active tunnel registrations.

**Independent Test**: Interrupt client connectivity or restart the relay, then confirm that the client reconnects and active tunnels become reachable again automatically.

### Tests for User Story 3

- [ ] T042 [P] [US3] Add integration test for client reconnect and tunnel re-registration in `/root/.openclaw/workspace/bloop-tunnel/test/integration/reconnect_reregister_test.go`
- [ ] T043 [P] [US3] Add integration test for relay restart recovery in `/root/.openclaw/workspace/bloop-tunnel/test/integration/relay_restart_recovery_test.go`

### Implementation for User Story 3

- [ ] T044 [P] [US3] Implement reconnect backoff policy in `/root/.openclaw/workspace/bloop-tunnel/internal/client/reconnect.go`
- [ ] T045 [US3] Implement heartbeat / ping-pong session liveness handling in `/root/.openclaw/workspace/bloop-tunnel/internal/transport/session.go`
- [ ] T046 [US3] Implement client-side active tunnel replay on reconnect in `/root/.openclaw/workspace/bloop-tunnel/internal/client/register.go`
- [ ] T047 [US3] Implement relay-side cleanup for dead sessions and hostname lease release in `/root/.openclaw/workspace/bloop-tunnel/internal/relay/session/manager.go`
- [ ] T048 [US3] Add structured logs for disconnect, reconnect, and re-registration events in `/root/.openclaw/workspace/bloop-tunnel/internal/logging/logging.go`

**Checkpoint**: All user stories should now be functional, including recovery from disconnects and relay restarts.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Deployment polish, operator docs, and final hardening across all stories.

- [ ] T049 [P] Add relay Dockerfile in `/root/.openclaw/workspace/bloop-tunnel/deploy/docker/relay.Dockerfile`
- [ ] T050 [P] Add client Dockerfile in `/root/.openclaw/workspace/bloop-tunnel/deploy/docker/client.Dockerfile`
- [ ] T051 [P] Add VPS relay Docker Compose example in `/root/.openclaw/workspace/bloop-tunnel/deploy/compose/relay-compose.yml`
- [ ] T052 [P] Add example client Compose file in `/root/.openclaw/workspace/bloop-tunnel/deploy/compose/example-client-compose.yml`
- [ ] T053 [P] Add Caddy example config in `/root/.openclaw/workspace/bloop-tunnel/deploy/examples/Caddyfile`
- [ ] T054 [P] Add Traefik example config in `/root/.openclaw/workspace/bloop-tunnel/deploy/examples/traefik-dynamic.yml`
- [ ] T055 Update `/root/.openclaw/workspace/bloop-tunnel/specs/001-v1-http-tunnels/quickstart.md` with end-to-end setup, DNS assumptions, and verification steps
- [ ] T056 [P] Add end-to-end smoke script or validation instructions in `/root/.openclaw/workspace/bloop-tunnel/scripts/smoke.sh`
- [ ] T057 Run full test suite and document known limitations in `/root/.openclaw/workspace/bloop-tunnel/README.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: starts immediately
- **Phase 2 (Foundational)**: depends on Setup completion and blocks all story work
- **Phase 3 (US1)**: starts after Foundational
- **Phase 4 (US2)**: starts after Foundational; integrates with US1 ingress path
- **Phase 5 (US3)**: starts after Foundational; depends on session/registration flow from US1
- **Phase 6 (Polish)**: depends on desired stories being complete

### User Story Dependencies

- **US1 (P1)**: no dependency on other stories after Foundational
- **US2 (P2)**: depends on US1 relay ingress path and tunnel model existing
- **US3 (P3)**: depends on US1 relay/client session flow and registration lifecycle

### Within Each User Story

- Write tests first where practical
- Implement core models/state before handlers
- Implement handlers before CLI wiring
- Validate each story independently before moving on

### Parallel Opportunities

- Setup tasks marked `[P]` can run in parallel
- Foundational tasks marked `[P]` can run in parallel
- US1 registry/router/client pieces marked `[P]` can be split initially
- US2 tests and access policy helpers can run in parallel
- US3 tests and reconnect primitives can run in parallel
- Docker/deploy example tasks in Phase 6 can run in parallel

---

## Implementation Strategy

### MVP First (Recommended)

1. Complete Setup
2. Complete Foundational phase
3. Complete User Story 1 only
4. Validate end-to-end hostname forwarding
5. Then add access control (US2)
6. Then add resilience (US3)

### Incremental Delivery

1. Foundation ready
2. Deliver public HTTP tunnel MVP
3. Add protected tunnels
4. Add reconnect/recovery
5. Add deployment polish and operator documentation

---

## Notes

- Keep V1 disciplined: HTTP/HTTPS only
- Prefer simple in-memory relay state before introducing persistence
- Avoid UI work unless a later spec explicitly adds it
- Keep config and deploy examples clean enough for eventual open-source publication
