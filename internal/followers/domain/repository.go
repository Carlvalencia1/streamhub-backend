package domain

import "context"

type Repository interface {
	Follow(ctx context.Context, followerID, streamerID string) error
	Unfollow(ctx context.Context, followerID, streamerID string) error
	IsFollowing(ctx context.Context, followerID, streamerID string) (bool, error)
	GetFollowerCount(ctx context.Context, streamerID string) (int, error)
	GetFollowingIDs(ctx context.Context, followerID string) ([]string, error)
}
