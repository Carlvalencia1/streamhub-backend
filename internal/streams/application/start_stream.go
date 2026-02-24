package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type StartStream struct {
	repo domain.Repository
}

func NewStartStream(repo domain.Repository) *StartStream {
	return &StartStream{repo: repo}
}

func (uc *StartStream) Execute(ctx context.Context, streamID string) error {
	return uc.repo.StartStream(ctx, streamID)
}