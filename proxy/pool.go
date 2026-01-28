package main

import (
	"log"
	"sync/atomic"
)

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func NewServerPool() *ServerPool {
	return &ServerPool{
		backends: make([]*Backend, 0),
		current:  0,
	}
}

func (p *ServerPool) AddBackend(backend *Backend) {
	p.backends = append(p.backends, backend)
}

func (p *ServerPool) GetBackends() []*Backend {
	return p.backends
}

func (p *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&p.current, 1) % uint64(len(p.backends)))
}

func (p *ServerPool) GetNextPeer() *Backend {
	next := p.NextIndex()
	l := len(p.backends) + next

	for i := next; i < l; i++ {
		idx := i % len(p.backends)
		if p.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&p.current, uint64(idx))
			}
			return p.backends[idx]
		}
	}
	return nil
}

func (p *ServerPool) MarkBackendAlive(backend *Backend, alive bool) {
	backend.SetAlive(alive)
	status := "ALIVE"
	if !alive {
		status = "DEAD"
	}
	log.Printf("Health check: %s is [%s]", backend, status)
}
