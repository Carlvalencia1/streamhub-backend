package domain

import "time"

type PostType string

const (
	PostTypeText       PostType = "text"
	PostTypeImage      PostType = "image"
	PostTypeVideo      PostType = "video"
	PostTypeShortVideo PostType = "short_video"
	PostTypeAudio      PostType = "audio"
	PostTypePoll       PostType = "poll"
)

type Post struct {
	ID         string    `json:"id"`
	StreamerID string    `json:"streamer_id"`
	Username   string    `json:"username"`
	AvatarURL  *string   `json:"avatar_url"`
	Type       PostType  `json:"type"`
	Content    string    `json:"content"`
	MediaURL   *string   `json:"media_url"`
	PollID     *string   `json:"poll_id"`
	Poll       *Poll     `json:"poll"`
	CreatedAt  time.Time `json:"created_at"`
}

type Poll struct {
	ID             string     `json:"id"`
	Question       string     `json:"question"`
	Options        []string   `json:"options"`
	MultipleChoice bool       `json:"multiple_choice"`
	Votes          []PollVote `json:"votes"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type PollVote struct {
	PollID      string `json:"poll_id"`
	UserID      string `json:"user_id"`
	OptionIndex int    `json:"option_index"`
}
