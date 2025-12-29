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
	"strings"
	"syscall"
)

const defaultPort = "8083"

var forbiddenWords = []string{"qwerty", "йцукен", "zxvbnm"}

func main() {
	port := flag.String("port", defaultPort, "HTTP server port")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", handleValidate)

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

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	text := strings.ToLower(req.Text)
	for _, word := range forbiddenWords {
		if strings.Contains(text, strings.ToLower(word)) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ValidateResponse{
				Valid: false,
				Error: fmt.Sprintf("Comment contains forbidden word: %s", word),
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateResponse{
		Valid: true,
	})
}

