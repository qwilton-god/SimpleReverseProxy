package proxy

import (
	"net/url"
	"sync"
)

type Backend struct {
	url   *url.URL
	alive bool
	mux   sync.RWMutex
}

func NewBackend(rawURL string) (*Backend, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &Backend{
		url:   parsedURL,
		alive: true,
	}, nil
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

func (b *Backend) String() string {
	return b.url.String()
}
