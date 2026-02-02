package proxy

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	url        *url.URL
	alive      bool
	weight     int
	currWeight atomic.Int64
	mux        sync.RWMutex
}

func NewBackend(rawURL string, weight int) (*Backend, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if weight <= 0 {
		weight = 1
	}

	b := &Backend{
		url:    parsedURL,
		alive:  true,
		weight: weight,
	}
	b.currWeight.Store(0)

	return b, nil
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.alive
}

func (b *Backend) GetURL() *url.URL {
	return b.url
}

func (b *Backend) GetStringURL() string {
	return b.url.String()
}

func (b *Backend) GetWeight() int {
	return b.weight
}

func (b *Backend) GetCurrentWeight() int64 {
	return b.currWeight.Load()
}

func (b *Backend) AddCurrentWeight(delta int64) int64 {
	return b.currWeight.Add(delta)
}
