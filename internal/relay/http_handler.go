package relay

import (
	"net/http"

	"bloop-relay/internal/relay/access"
	"bloop-relay/internal/relay/routing"
)

type HTTPHandler struct {
	Router *routing.Router
	Proxy  *RequestProxy
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tunnel, ok := h.Router.Resolve(r.Host)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if status, msg, allowed := access.EnforceTunnel(r, tunnel); !allowed {
		if status == http.StatusUnauthorized && access.ResolveMode(tunnel.Access) == access.ModeBasicAuth {
			w.Header().Set("WWW-Authenticate", `Basic realm="bloop-relay"`)
		}
		http.Error(w, msg, status)
		return
	}

	if err := h.Proxy.ForwardHTTP(w, r, tunnel.SessionID); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
}
