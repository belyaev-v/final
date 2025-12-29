package main

import (
	"database/sql"
	"fmt"

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
	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		news_id INTEGER NOT NULL,
		text TEXT NOT NULL,
		parent_comment_id INTEGER,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	`
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create comments table: %w", err)
	}

	return nil
}

func (db *DB) CreateComment(newsID int, text string, parentCommentID *int) (*Comment, error) {
	var comment Comment
	var parentID sql.NullInt64
	if parentCommentID != nil {
		parentID = sql.NullInt64{Int64: int64(*parentCommentID), Valid: true}
	}

	err := db.conn.QueryRow(
		"INSERT INTO comments (news_id, text, parent_comment_id, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id, news_id, text, parent_comment_id, created_at",
		newsID, text, parentID,
	).Scan(&comment.ID, &comment.NewsID, &comment.Text, &parentID, &comment.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	if parentID.Valid {
		parentIDInt := int(parentID.Int64)
		comment.ParentCommentID = &parentIDInt
	}

	return &comment, nil
}

func (db *DB) GetCommentsByNewsID(newsID int) ([]Comment, error) {
	rows, err := db.conn.Query(
		"SELECT id, news_id, text, parent_comment_id, created_at FROM comments WHERE news_id = $1 ORDER BY created_at ASC",
		newsID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var parentID sql.NullInt64
		if err := rows.Scan(&comment.ID, &comment.NewsID, &comment.Text, &parentID, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		if parentID.Valid {
			parentIDInt := int(parentID.Int64)
			comment.ParentCommentID = &parentIDInt
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

