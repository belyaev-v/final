package main

import "time"

type NewsShortDetailed struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	PubTime time.Time `json:"pub_time"`
}

type NewsFullDetailed struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	PubTime   time.Time `json:"pub_time"`
	Link      string    `json:"link"`
	Source    string    `json:"source"`
	Comments  []Comment `json:"comments"`
}

type Comment struct {
	ID              int       `json:"id"`
	NewsID          int       `json:"news_id"`
	Text            string    `json:"text"`
	ParentCommentID *int      `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type NewsListResponse struct {
	News  []NewsShortDetailed `json:"news"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Pages int                 `json:"pages"`
}

type CreateCommentRequest struct {
	Text            string `json:"text"`
	ParentCommentID *int   `json:"parent_comment_id,omitempty"`
}

