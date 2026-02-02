package proxy

import (
	"log/slog"
	"net/http"
)

type LoadBalancer struct {
	pool  *ServerPool
	proxy *Proxy
}

func NewLoadBalancer(pool *ServerPool) *LoadBalancer {
	return &LoadBalancer{
		pool:  pool,
		proxy: NewProxy(),
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := lb.pool.GetNextPeer()
	if peer == nil {
		slog.Error("No available backends",
			"method", r.Method,
			"path", r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	slog.Debug("Proxying request",
		"method", r.Method,
		"path", r.URL.Path,
		"backend", peer.GetStringURL())

	lb.proxy.ServeHTTP(w, r, peer)
}
