package proxy

import (
	"log/slog"
	"sync"
)

type ServerPool struct {
	backends []*Backend
	mu       sync.RWMutex
}

func NewServerPool() *ServerPool {
	return &ServerPool{
		backends: make([]*Backend, 0),
	}
}

func (p *ServerPool) AddBackend(backend *Backend) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.backends = append(p.backends, backend)
}

func (p *ServerPool) GetBackends() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.backends
}


func (p *ServerPool) GetNextPeer() *Backend {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.backends) == 0 {
		return nil
	}

	total := int64(0)
	for _, backend := range p.backends {
		if !backend.IsAlive() {
			continue
		}
		weight := int64(backend.GetWeight())
		total += weight
		backend.AddCurrentWeight(weight)
	}

	if total == 0 {
		return nil
	}


	var selected *Backend
	var maxWeight int64 = -1

	for _, backend := range p.backends {
		if !backend.IsAlive() {
			continue
		}

		currWeight := backend.GetCurrentWeight()
		if currWeight > maxWeight {
			maxWeight = currWeight
			selected = backend
		}
	}

	if selected == nil {
		return nil
	}


	selected.AddCurrentWeight(-total)

	return selected
}

func (p *ServerPool) MarkBackendAlive(backend *Backend, alive bool) {
	backend.SetAlive(alive)

	if alive {
		slog.Debug("Backend is alive", "url", backend.GetStringURL())
	} else {
		slog.Warn("Backend is dead", "url", backend.GetStringURL())
	}
}
