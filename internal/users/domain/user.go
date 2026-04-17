package domain

import "time"

type User struct {
	ID        string
	Username  string
	Email     string
	Password  string
	Role      string
	Nickname  *string
	Bio       *string
	Location  *string
	AvatarURL *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
