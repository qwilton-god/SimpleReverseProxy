package main

import (
	"log"
	"net"
	"net/url"
	"time"
)

const (
	defaultHealthCheckTimeout = 2 * time.Second
)

func (p *ServerPool) HealthCheck() {
	for _, backend := range p.backends {
		alive := isBackendAlive(backend.GetURL())
		p.MarkBackendAlive(backend, alive)
	}
}

func isBackendAlive(u *url.URL) bool {
	conn, err := net.DialTimeout("tcp", u.Host, defaultHealthCheckTimeout)
	if err != nil {
		log.Printf("Health check failed for %s: %v", u, err)
		return false
	}
	_ = conn.Close()
	return true
}
