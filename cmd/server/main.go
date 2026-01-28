package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reverseProxyBasic/pkg/config"
)

type SimpleServer struct {
	port string
	name string
}

func NewSimpleServer(port, name string) *SimpleServer {
	return &SimpleServer{port: port, name: name}
}

func (s *SimpleServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	response := fmt.Sprintf("[%s] Server: %s | Path: %s | Method: %s\n",
		timestamp, s.name, r.URL.Path, r.Method)
	w.Write([]byte(response))
	log.Printf("%s -> %s", s.name, r.URL.Path)
}

func main() {
	cfg := config.LoadMust()

	var servers []*http.Server

	for i, port := range cfg.ServerPorts {
		srv := NewSimpleServer(port, fmt.Sprintf("Server-%d", i+1))

		server := &http.Server{
			Addr:    port,
			Handler: srv,
		}

		servers = append(servers, server)

		go func(s *SimpleServer, addr string) {
			log.Printf("Starting %s on %s", s.name, addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("%s failed: %v", s.name, err)
			}
		}(srv, port)
	}

	log.Println("All servers started. Press Ctrl+C to stop.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Servers shutting down...")

	for _, srv := range servers {
		_ = srv.Close()
	}

	log.Println("All servers stopped")
}
