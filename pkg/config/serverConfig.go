package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerConfig struct {
	ServerPorts []string `json:"server_ports"`
}

func SLoad() (*ServerConfig, error) {
	data, err := os.ReadFile("server.json")
	if err != nil {
		return nil, err
	}
	var cfg ServerConfig

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.ServerPorts) == 0 {
		return nil, ErrNoServers
	}

	return &cfg, nil
}

func ServerLoadMust() *ServerConfig {
	cfg, err := SLoad()
	if err != nil {
		panic(fmt.Sprintf("config validation failed: %v", err))
	}

	return cfg
}
