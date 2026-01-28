package main

import (
	"log"
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
		log.Printf("No available backends for request %s %s", r.Method, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Proxying %s %s to backend %s", r.Method, r.URL.Path, peer)
	lb.proxy.ServeHTTP(w, r, peer)
}
