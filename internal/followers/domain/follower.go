package domain

import "time"

type Follower struct {
	ID         string
	FollowerID string
	StreamerID string
	CreatedAt  time.Time
}
