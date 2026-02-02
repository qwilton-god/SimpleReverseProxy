package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reverseProxyBasic/pkg/config"
	"reverseProxyBasic/pkg/proxy"
)

func main() {
	// Configure structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.LoadMust()

	pool := proxy.NewServerPool()

	for _, server := range cfg.Servers {
		backend, err := proxy.NewBackend(server.URL, server.Weight)
		if err != nil {
			slog.Error("Failed to create backend",
				"url", server.URL,
				"error", err)
			os.Exit(1)
		}

		pool.AddBackend(backend)
		slog.Info("Configured backend",
			"url", backend.GetStringURL(),
			"weight", server.Weight)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.HealthCheckInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			pool.HealthCheck()
		}
	}()

	lb := proxy.NewLoadBalancer(pool)

	server := &http.Server{
		Addr:         cfg.ProxyPort,
		Handler:      lb,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Load balancer started", "addr", cfg.ProxyPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited successfully")
}
