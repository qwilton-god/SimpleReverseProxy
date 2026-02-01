package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var (
	ErrInvalidPort     = errors.New("invalid port format")
	ErrInvalidInterval = errors.New("invalid health check interval")
	ErrNoServers       = errors.New("no backends provided in config")
)

type Config struct {
	ProxyPort           string   `json:"proxy_port"`
	Servers             []string `json:"servers"`
	HealthCheckInterval int      `json:"health_check_interval"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var cfg Config

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Servers) == 0 {
		return nil, ErrNoServers
	}
	if cfg.ProxyPort == "" {
		return nil, ErrInvalidPort
	}
	if cfg.HealthCheckInterval <= 0 {
		return nil, ErrInvalidInterval
	}

	return &cfg, nil
}

func LoadMust() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config validation failed: %v", err))
	}

	return cfg
}
