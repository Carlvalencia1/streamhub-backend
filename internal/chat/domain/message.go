package domain

import "time"

type Message struct {
	ID        string
	StreamID  string
	UserID    string
	Content   string
	CreatedAt time.Time
}