package routing

import "bloop-tunnel/internal/relay/registry"

type Router struct {
	registry *registry.Registry
}

func New(reg *registry.Registry) *Router {
	return &Router{registry: reg}
}

func (r *Router) Resolve(host string) (registry.Tunnel, bool) {
	return r.registry.Lookup(host)
}
