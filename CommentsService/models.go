package main

import "time"

type Comment struct {
	ID              int       `json:"id"`
	NewsID          int       `json:"news_id"`
	Text            string    `json:"text"`
	ParentCommentID *int      `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateCommentRequest struct {
	NewsID          int    `json:"news_id"`
	Text            string `json:"text"`
	ParentCommentID *int   `json:"parent_comment_id,omitempty"`
}

