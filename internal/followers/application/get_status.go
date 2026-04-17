package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/followers/domain"
)

type GetFollowerStatus struct {
	repo domain.Repository
}

func NewGetFollowerStatus(repo domain.Repository) *GetFollowerStatus {
	return &GetFollowerStatus{repo: repo}
}

func (uc *GetFollowerStatus) Execute(ctx context.Context, followerID, streamerID string) (bool, int, error) {
	isFollowing, err := uc.repo.IsFollowing(ctx, followerID, streamerID)
	if err != nil {
		return false, 0, err
	}
	count, err := uc.repo.GetFollowerCount(ctx, streamerID)
	return isFollowing, count, err
}

type GetFollowing struct {
	repo domain.Repository
}

func NewGetFollowing(repo domain.Repository) *GetFollowing {
	return &GetFollowing{repo: repo}
}

func (uc *GetFollowing) Execute(ctx context.Context, followerID string) ([]string, error) {
	return uc.repo.GetFollowingIDs(ctx, followerID)
}
