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
	"strconv"
	"syscall"
)

const defaultPort = "8082"
const defaultDSN = "postgres://postgres:postgres@localhost:5432/commentsdb?sslmode=disable"

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
	mux.HandleFunc("/comments", handleComments)

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	slog.Info("Shutting down server...")
	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
}

func handleComments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateComment(w, r)
	case http.MethodGet:
		handleGetComments(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCreateComment(w http.ResponseWriter, r *http.Request) {
	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	if req.NewsID == 0 {
		http.Error(w, "NewsID is required", http.StatusBadRequest)
		return
	}

	comment, err := db.CreateComment(req.NewsID, req.Text, req.ParentCommentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create comment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func handleGetComments(w http.ResponseWriter, r *http.Request) {
	newsIDStr := r.URL.Query().Get("news_id")
	if newsIDStr == "" {
		http.Error(w, "news_id parameter is required", http.StatusBadRequest)
		return
	}

	newsID, err := strconv.Atoi(newsIDStr)
	if err != nil {
		http.Error(w, "Invalid news_id", http.StatusBadRequest)
		return
	}

	comments, err := db.GetCommentsByNewsID(newsID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get comments: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
