package main

import (
	"fmt"
	"syscall"
)

type Config struct {
	ProxyPort           string
	ServerPorts         []string
	HealthCheckInterval int
}

func getEnv(key, fallback string) string {
	if value, exists := syscall.Getenv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := syscall.Getenv(key); exists {
		var intValue int
		_, err := fmt.Sscanf(value, "%d", &intValue)
		if err == nil {
			return intValue
		}
	}
	return fallback
}

func Load() *Config {
	serverPorts := []string{
		getEnv("SERVER_1_PORT", ":8081"),
		getEnv("SERVER_2_PORT", ":8082"),
		getEnv("SERVER_3_PORT", ":8083"),
	}

	return &Config{
		ProxyPort:           getEnv("PROXY_PORT", ":8080"),
		ServerPorts:         serverPorts,
		HealthCheckInterval: getEnvInt("HEALTH_CHECK_INTERVAL", 10),
	}
}
