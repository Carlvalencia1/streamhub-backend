package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type StopStream struct {
	repo domain.Repository
}

func NewStopStream(repo domain.Repository) *StopStream {
	return &StopStream{repo: repo}
}

func (uc *StopStream) Execute(ctx context.Context, streamID string) error {
	return uc.repo.StopStream(ctx, streamID)
}
