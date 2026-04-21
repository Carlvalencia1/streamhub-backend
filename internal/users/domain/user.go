package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	Role      string    `json:"role"`
	Nickname  *string   `json:"nickname"`
	Bio       *string   `json:"bio"`
	Location  *string   `json:"location"`
	AvatarURL *string   `json:"avatar_url"`
	BannerURL *string   `json:"banner_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
