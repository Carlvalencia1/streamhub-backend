package domain

import "time"

type Follower struct {
	ID         string    `json:"id"`
	FollowerID string    `json:"follower_id"`
	StreamerID string    `json:"streamer_id"`
	CreatedAt  time.Time `json:"created_at"`
}
