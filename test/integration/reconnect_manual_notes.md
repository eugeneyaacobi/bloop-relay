# Manual reconnect verification notes

Expected reconnect behavior for MVP:
- client notices relay disconnect
- client waits according to reconnect policy
- client reconnects using the same config
- client re-registers all configured tunnels
- relay accepts the new session and hostname leases
- public requests succeed again without manual reconfiguration
