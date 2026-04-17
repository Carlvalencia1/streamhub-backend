package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/followers/domain"
)

type Follow struct {
	repo domain.Repository
}

func NewFollow(repo domain.Repository) *Follow {
	return &Follow{repo: repo}
}

func (uc *Follow) Execute(ctx context.Context, followerID, streamerID string) error {
	if followerID == streamerID {
		return domain.ErrCannotFollowSelf
	}
	return uc.repo.Follow(ctx, followerID, streamerID)
}
