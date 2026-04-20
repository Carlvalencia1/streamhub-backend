package domain

import "context"

type UserSummary struct {
	ID        string
	Username  string
	Nickname  *string
	AvatarURL *string
}

type Repository interface {
	Follow(ctx context.Context, followerID, streamerID string) error
	Unfollow(ctx context.Context, followerID, streamerID string) error
	IsFollowing(ctx context.Context, followerID, streamerID string) (bool, error)
	GetFollowerCount(ctx context.Context, streamerID string) (int, error)
	GetFollowingIDs(ctx context.Context, followerID string) ([]string, error)
	GetFollowerUsers(ctx context.Context, streamerID string) ([]*UserSummary, error)
	GetFollowingUsers(ctx context.Context, followerID string) ([]*UserSummary, error)
}
