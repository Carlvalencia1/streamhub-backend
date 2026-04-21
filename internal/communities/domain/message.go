package domain

import "time"

type MessageType string

const (
	MessageTypeText       MessageType = "text"
	MessageTypeImage      MessageType = "image"
	MessageTypeVideo      MessageType = "video"
	MessageTypeAudio      MessageType = "audio"
	MessageTypePoll       MessageType = "poll"
	MessageTypeShortVideo MessageType = "short_video"
)

type ChannelMessage struct {
	ID         string      `json:"id"`
	ChannelID  string      `json:"channel_id"`
	UserID     string      `json:"user_id"`
	Username   string      `json:"username"`
	AvatarURL  *string     `json:"avatar_url"`
	Type       MessageType `json:"type"`
	Content    string      `json:"content"`
	MediaURL   *string     `json:"media_url"`
	PollID     *string     `json:"poll_id"`
	Poll       *Poll       `json:"poll"`
	MyReaction *string     `json:"my_reaction"`
	ExpiresAt  *time.Time  `json:"expires_at"`
	CreatedAt  time.Time   `json:"created_at"`
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

type ChannelSettings struct {
	ChannelID              string `json:"channel_id"`
	DisappearingTTLSeconds int    `json:"disappearing_ttl_seconds"`
}
