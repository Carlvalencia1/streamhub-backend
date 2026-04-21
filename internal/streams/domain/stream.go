package domain

import "time"

type Stream struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ThumbnailURL string     `json:"thumbnail_url"`
	Category     string     `json:"category"`
	OwnerID      string     `json:"owner_id"`
	ViewersCount int        `json:"viewers_count"`
	IsLive       bool       `json:"is_live"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	StreamKey    string     `json:"stream_key"`
	PlaybackURL  string     `json:"playback_url"`
	RecordingURL *string    `json:"recording_url,omitempty"`
}