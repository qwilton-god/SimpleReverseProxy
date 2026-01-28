package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrInvalidPort   = errors.New("invalid port format")
	ErrInvalidInterval = errors.New("invalid health check interval")
)

type Config struct {
	ProxyPort           string
	ServerPorts         []string
	HealthCheckInterval int
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func validatePort(port string) error {
	if !strings.HasPrefix(port, ":") {
		return fmt.Errorf("%w: port must start with ':', got: %s", ErrInvalidPort, port)
	}
	portNum, err := strconv.Atoi(port[1:])
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidPort, port)
	}
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("%w: port out of range: %s", ErrInvalidPort, port)
	}
	return nil
}

func (c *Config) Validate() error {
	if err := validatePort(c.ProxyPort); err != nil {
		return fmt.Errorf("proxy port: %w", err)
	}

	if len(c.ServerPorts) == 0 {
		return errors.New("at least one server port is required")
	}

	for i, port := range c.ServerPorts {
		if err := validatePort(port); err != nil {
			return fmt.Errorf("server port %d: %w", i+1, err)
		}
	}

	if c.HealthCheckInterval < 1 {
		return fmt.Errorf("%w: must be at least 1 second", ErrInvalidInterval)
	}

	return nil
}

func Load() (*Config, error) {
	cfg := &Config{
		ProxyPort:           getEnv("PROXY_PORT", ":8080"),
		ServerPorts: []string{
			getEnv("SERVER_1_PORT", ":8081"),
			getEnv("SERVER_2_PORT", ":8082"),
			getEnv("SERVER_3_PORT", ":8083"),
		},
		HealthCheckInterval: getEnvInt("HEALTH_CHECK_INTERVAL", 10),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func LoadMust() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config validation failed: %v", err))
	}
	return cfg
}
