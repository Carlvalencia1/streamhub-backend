package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type JoinStream struct {
	repo domain.Repository
}

func NewJoinStream(repo domain.Repository) *JoinStream {
	return &JoinStream{repo: repo}
}

func (uc *JoinStream) Execute(ctx context.Context, streamID string) error {
	return uc.repo.JoinStream(ctx, streamID)
}