package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS news (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		pub_time TIMESTAMP NOT NULL,
		link TEXT,
		source TEXT
	);
	`
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create news table: %w", err)
	}

	// Добавим тестовые данные, если таблица пустая
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM news").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count news: %w", err)
	}

	if count == 0 {
		log.Println("Adding sample news data...")
		sampleNews := []struct {
			title, content, link, source string
		}{
			{"Новость о Go", "Go - отличный язык программирования", "https://example.com/go", "Example"},
			{"Новость о микросервисах", "Микросервисная архитектура становится популярной", "https://example.com/microservices", "Example"},
			{"Новость о PostgreSQL", "PostgreSQL - мощная реляционная БД", "https://example.com/postgres", "Example"},
			{"Новость о Docker", "Docker упрощает развертывание приложений", "https://example.com/docker", "Example"},
			{"Новость о Kubernetes", "Kubernetes для оркестрации контейнеров", "https://example.com/k8s", "Example"},
		}

		for _, n := range sampleNews {
			_, err = db.conn.Exec(
				"INSERT INTO news (title, content, pub_time, link, source) VALUES ($1, $2, NOW(), $3, $4)",
				n.title, n.content, n.link, n.source,
			)
			if err != nil {
				return fmt.Errorf("failed to insert sample news: %w", err)
			}
		}
	}

	return nil
}

func (db *DB) GetNews(page, pageSize int, search string) ([]NewsShortDetailed, int, error) {
	offset := (page - 1) * pageSize
	var rows *sql.Rows
	var err error

	if search != "" {
		query := `
			SELECT id, title, content, pub_time 
			FROM news 
			WHERE title ILIKE $1 
			ORDER BY pub_time DESC 
			LIMIT $2 OFFSET $3
		`
		rows, err = db.conn.Query(query, "%"+search+"%", pageSize, offset)
	} else {
		query := `
			SELECT id, title, content, pub_time 
			FROM news 
			ORDER BY pub_time DESC 
			LIMIT $1 OFFSET $2
		`
		rows, err = db.conn.Query(query, pageSize, offset)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to query news: %w", err)
	}
	defer rows.Close()

	var news []NewsShortDetailed
	for rows.Next() {
		var n NewsShortDetailed
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.PubTime); err != nil {
			return nil, 0, fmt.Errorf("failed to scan news: %w", err)
		}
		news = append(news, n)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM news"
	if search != "" {
		countQuery = "SELECT COUNT(*) FROM news WHERE title ILIKE $1"
		err = db.conn.QueryRow(countQuery, "%"+search+"%").Scan(&total)
	} else {
		err = db.conn.QueryRow(countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count news: %w", err)
	}

	return news, total, nil
}

func (db *DB) GetNewsByID(id int) (*NewsFullDetailed, error) {
	var news NewsFullDetailed
	err := db.conn.QueryRow(
		"SELECT id, title, content, pub_time, link, source FROM news WHERE id = $1",
		id,
	).Scan(&news.ID, &news.Title, &news.Content, &news.PubTime, &news.Link, &news.Source)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}

	return &news, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

