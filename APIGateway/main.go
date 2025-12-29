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
	"sync"
	"syscall"
)

const defaultPort = "8080"
const defaultNewsServiceURL = "http://localhost:8081"
const defaultCommentsServiceURL = "http://localhost:8082"
const defaultCensorshipServiceURL = "http://localhost:8083"

var (
	newsServiceClient       *HTTPClient
	commentsServiceClient   *HTTPClient
	censorshipServiceClient *HTTPClient
)

func main() {
	port := flag.String("port", defaultPort, "HTTP server port")
	newsURL := flag.String("news-url", defaultNewsServiceURL, "News service URL")
	commentsURL := flag.String("comments-url", defaultCommentsServiceURL, "Comments service URL")
	censorshipURL := flag.String("censorship-url", defaultCensorshipServiceURL, "Censorship service URL")
	flag.Parse()

	newsServiceClient = NewHTTPClient(*newsURL)
	commentsServiceClient = NewHTTPClient(*commentsURL)
	censorshipServiceClient = NewHTTPClient(*censorshipURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/news", handleNews)
	mux.HandleFunc("/news/filter", handleFilterNews)
	mux.HandleFunc("/news/", handleNewsByID)

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

func handleNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestID := r.Header.Get("X-Request-ID")
	query := r.URL.Query()

	path := "/news"
	if s := query.Get("s"); s != "" {
		path += "?s=" + s
	}
	if page := query.Get("page"); page != "" {
		if path == "/news" {
			path += "?page=" + page
		} else {
			path += "&page=" + page
		}
	}

	resp, err := newsServiceClient.Get(path, requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get news: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := readResponseBody(resp)
		http.Error(w, string(body), resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	var newsResponse NewsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newsResponse)
}

func handleFilterNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestID := r.Header.Get("X-Request-ID")
	query := r.URL.Query()

	path := "/news"
	if s := query.Get("s"); s != "" {
		path += "?s=" + s
	}
	if page := query.Get("page"); page != "" {
		if path == "/news" {
			path += "?page=" + page
		} else {
			path += "&page=" + page
		}
	}

	resp, err := newsServiceClient.Get(path, requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to filter news: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := readResponseBody(resp)
		http.Error(w, string(body), resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	var newsResponse NewsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newsResponse)
}

func handleNewsByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	// Проверяем, является ли это запросом на создание комментария
	if r.Method == http.MethodPost {
		var id int
		_, err := fmt.Sscanf(path, "/news/%d/comments", &id)
		if err == nil {
			handleCreateComment(w, r, id)
			return
		}
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	// Иначе это запрос на получение новости по ID
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id int
	_, err := fmt.Sscanf(path, "/news/%d", &id)
	if err != nil {
		http.Error(w, "Invalid news ID", http.StatusBadRequest)
		return
	}

	requestID := r.Header.Get("X-Request-ID")

	// Асинхронное получение данных из двух сервисов
	type newsResult struct {
		news *NewsFullDetailed
		err  error
	}
	type commentsResult struct {
		comments []Comment
		err      error
	}

	newsChan := make(chan newsResult, 1)
	commentsChan := make(chan commentsResult, 1)
	var wg sync.WaitGroup

	// Получение новости
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := newsServiceClient.Get(fmt.Sprintf("/news/%d", id), requestID)
		if err != nil {
			newsChan <- newsResult{nil, err}
			return
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := readResponseBody(resp)
			newsChan <- newsResult{nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))}
			return
		}
		defer resp.Body.Close()

		var news NewsFullDetailed
		if err := json.NewDecoder(resp.Body).Decode(&news); err != nil {
			newsChan <- newsResult{nil, err}
			return
		}
		newsChan <- newsResult{&news, nil}
	}()

	// Получение комментариев
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := commentsServiceClient.Get(fmt.Sprintf("/comments?news_id=%d", id), requestID)
		if err != nil {
			commentsChan <- commentsResult{nil, err}
			return
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := readResponseBody(resp)
			commentsChan <- commentsResult{nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))}
			return
		}
		defer resp.Body.Close()

		var comments []Comment
		if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
			commentsChan <- commentsResult{nil, err}
			return
		}
		commentsChan <- commentsResult{comments, nil}
	}()

	wg.Wait()

	newsRes := <-newsChan
	commentsRes := <-commentsChan

	if newsRes.err != nil {
		http.Error(w, fmt.Sprintf("Failed to get news: %v", newsRes.err), http.StatusInternalServerError)
		return
	}

	if commentsRes.err != nil {
		http.Error(w, fmt.Sprintf("Failed to get comments: %v", commentsRes.err), http.StatusInternalServerError)
		return
	}

	newsRes.news.Comments = commentsRes.comments

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newsRes.news)
}

func handleCreateComment(w http.ResponseWriter, r *http.Request, newsID int) {
	requestID := r.Header.Get("X-Request-ID")

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Сначала проверяем через сервис цензуры
	validateReq := map[string]string{"text": req.Text}
	resp, err := censorshipServiceClient.Post("/validate", validateReq, requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to validate comment: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := readResponseBody(resp)
		http.Error(w, string(body), resp.StatusCode)
		return
	}
	resp.Body.Close()

	// Если валидация прошла, создаем комментарий
	createCommentReq := map[string]interface{}{
		"news_id": newsID,
		"text":    req.Text,
	}
	if req.ParentCommentID != nil {
		createCommentReq["parent_comment_id"] = *req.ParentCommentID
	}

	resp, err = commentsServiceClient.Post("/comments", createCommentReq, requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create comment: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := readResponseBody(resp)
		http.Error(w, string(body), resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	var comment Comment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

