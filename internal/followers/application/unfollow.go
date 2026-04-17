package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/followers/domain"
)

type Unfollow struct {
	repo domain.Repository
}

func NewUnfollow(repo domain.Repository) *Unfollow {
	return &Unfollow{repo: repo}
}

func (uc *Unfollow) Execute(ctx context.Context, followerID, streamerID string) error {
	return uc.repo.Unfollow(ctx, followerID, streamerID)
}
