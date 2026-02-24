package domain

import "time"

type Stream struct {
	ID           string
	Title        string
	Description  string
	ThumbnailURL string
	Category     string
	OwnerID      string
	ViewersCount int
	IsLive       bool
	StartedAt    *time.Time
	CreatedAt    time.Time
}