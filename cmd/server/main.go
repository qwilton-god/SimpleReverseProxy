package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reverseProxyBasic/pkg/config"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg := config.ServerLoadMust()

	var wg sync.WaitGroup
	servers := make([]*http.Server, 0, len(cfg.ServerPorts))

	for _, port := range cfg.ServerPorts {
		addr := ":" + port

		mux := http.NewServeMux()
		registerHandlers(mux, addr)

		srv := &http.Server{
			Addr:    addr,
			Handler: mux,
		}
		servers = append(servers, srv)

		wg.Add(1)
		go func(p string, s *http.Server) {
			defer wg.Done()
			log.Printf("[START] Test Server on %s", p)
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[ERROR] Server %s: %v", p, err)
			}
		}(port, srv)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, s := range servers {
		s.Shutdown(ctx)
	}
	log.Println("All stopped")
}

func registerHandlers(mux *http.ServeMux, serverAddr string) {

	// 1. Стандартный ответ
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s! You requested: %s\n", serverAddr, r.URL.Path)
	})

	// 2. Health Check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 3. Медленный ответ 
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprintf(w, "Slow response from %s\n", serverAddr)
	})

	// 4. Возврат ошибки
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error from %s\n", serverAddr)
	})

	// 5. Эхо заголовков
	mux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Headers received by %s:\n", serverAddr)
		for name, values := range r.Header {
			for _, v := range values {
				fmt.Fprintf(w, "%s: %s\n", name, v)
			}
		}
	})

	// 6. Обработка GET с Query-параметрами
	// Пример вызова: curl "http://localhost:8080/search?query=canon&limit=10000"
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query().Get("query")
		limit := r.URL.Query().Get("limit")

		fmt.Fprintf(w, "[%s] Search triggered!\n", serverAddr)
		fmt.Fprintf(w, "Query: %s, Limit: %s\n", query, limit)
	})

	// 7. Обработка POST запроса (чтение тела)
	// Пример вызова: curl -X POST -d "{\"price\": 100}" http://localhost:8080/echo
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		fmt.Fprintf(w, "[%s] POST received!\n", serverAddr)
		fmt.Fprintf(w, "Received Body: %s\n", string(body))
	})
}
