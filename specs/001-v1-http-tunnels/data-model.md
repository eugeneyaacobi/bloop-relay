# Data Model: bloop-tunnel v1 HTTP tunnel system

## Relay

Public service responsible for:
- authenticating clients
- maintaining active sessions
- tracking tunnel registrations
- enforcing access policy
- routing inbound requests by hostname

## Client Session

Represents a live authenticated connection from a client instance to the relay.

Fields:
- session_id
- client_identity
- connected_at
- last_seen_at
- connection_state
- registered_tunnel_ids

## Tunnel

Represents a single public hostname mapped to a local HTTP target.

Fields:
- tunnel_id
- session_id
- name
- hostname
- local_addr
- access_mode
- created_at
- status

## Access Policy

Represents how a tunnel is protected.

Fields:
- mode (`public`, `basic_auth`, `token_protected`)
- username (for basic auth)
- secret reference metadata

## Hostname Lease

Represents relay-side ownership of a hostname.

Fields:
- hostname
- tunnel_id
- session_id
- active
- created_at

## Forwarded Request

Represents a public request in transit between relay and client.

Fields:
- request_id
- hostname
- method
- path
- headers
- started_at
- session_id
- tunnel_id

## Forwarded Response

Represents the response returned by the local target through the client.

Fields:
- request_id
- status_code
- headers
- body_stream_state
- completed_at
