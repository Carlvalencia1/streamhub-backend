package domain

import "time"

type User struct {
	ID        string
	Username  string
	Email     string
	Password  string
	AvatarURL *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
