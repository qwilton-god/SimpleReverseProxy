package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
)

var (
	ErrInvalidPort     = errors.New("invalid port format")
	ErrInvalidInterval = errors.New("invalid health check interval")
	ErrNoServers       = errors.New("no backends provided in config")
	ErrInvalidURL      = errors.New("invalid server URL")
)

type Server struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

type serverConfig struct {
	ProxyPort           string   `json:"proxy_port"`
	Servers             []Server `json:"servers"`
	HealthCheckInterval int      `json:"health_check_interval"`
}

type Config struct {
	ProxyPort           string
	Servers             []Server
	HealthCheckInterval int
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var sc serverConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, err
	}

	if len(sc.Servers) == 0 {
		return nil, ErrNoServers
	}
	if sc.ProxyPort == "" {
		return nil, ErrInvalidPort
	}
	if sc.HealthCheckInterval <= 0 {
		return nil, ErrInvalidInterval
	}

	// Validate URLs and set default weights
	for i := range sc.Servers {
		if sc.Servers[i].URL == "" {
			return nil, fmt.Errorf("%w: server %d has empty URL", ErrInvalidURL, i)
		}
		if _, err := url.Parse(sc.Servers[i].URL); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidURL, sc.Servers[i].URL)
		}
		if sc.Servers[i].Weight <= 0 {
			sc.Servers[i].Weight = 1
		}
	}

	return &Config{
		ProxyPort:           sc.ProxyPort,
		Servers:             sc.Servers,
		HealthCheckInterval: sc.HealthCheckInterval,
	}, nil
}

func LoadMust() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config validation failed: %v", err))
	}

	return cfg
}
