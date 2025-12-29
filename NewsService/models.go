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
}

type NewsListResponse struct {
	News  []NewsShortDetailed `json:"news"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Pages int                 `json:"pages"`
}

