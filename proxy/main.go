package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := Load()

	pool := NewServerPool()

	for _, port := range cfg.ServerPorts {
		backend, err := NewBackend("http://localhost" + port)
		if err != nil {
			log.Fatalf("Failed to create backend for port %s: %v", port, err)
		}

		pool.AddBackend(backend)
		log.Printf("Configured backend: %s", backend)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.HealthCheckInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			pool.HealthCheck()
		}
	}()

	lb := NewLoadBalancer(pool)

	server := &http.Server{
		Addr:         cfg.ProxyPort,
		Handler:      lb,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Load Balancer started at %s", cfg.ProxyPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}
