package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const defaultPort = "8081"
const defaultDSN = "postgres://postgres:postgres@localhost:5433/newsdb?sslmode=disable"

var db *DB

func main() {
	port := flag.String("port", defaultPort, "HTTP server port")
	dsn := flag.String("dsn", defaultDSN, "Database connection string")
	flag.Parse()

	var err error
	db, err = NewDB(*dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/news", handleGetNews)
	mux.HandleFunc("/news/", handleGetNewsByID)

	handler := requestIDMiddleware(loggingMiddleware(mux))

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: handler,
	}

	go func() {
		slog.Info(fmt.Sprintf("[*] HTTP server is started on localhost:%s", *port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Gracefull shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	slog.Info("Shutting down server...")
	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
}

func handleGetNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}

	search := r.URL.Query().Get("s")
	pageSize := 10

	news, total, err := db.GetNews(page, pageSize, search)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get news: %v", err), http.StatusInternalServerError)
		return
	}

	pages := (total + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}

	response := NewsListResponse{
		News:  news,
		Total: total,
		Page:  page,
		Pages: pages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetNewsByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id int
	_, err := fmt.Sscanf(r.URL.Path, "/news/%d", &id)
	if err != nil {
		http.Error(w, "Invalid news ID", http.StatusBadRequest)
		return
	}

	news, err := db.GetNewsByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get news: %v", err), http.StatusInternalServerError)
		return
	}

	if news == nil {
		http.Error(w, "News not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(news)
}
