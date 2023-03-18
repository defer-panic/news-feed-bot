package model

import (
	"time"
)

type Item struct {
	Title       string
	Description string
	Link        string
	Date        time.Time
	SourceName  string
}

type Source struct {
	ID        int64
	Name      string
	FeedURL   string
	Priority  int
	CreatedAt time.Time
}

type Article struct {
	ID          int64
	SourceID    int64
	Title       string
	Link        string
	PublishedAt time.Time
	PostedAt    time.Time
	CreatedAt   time.Time
}
