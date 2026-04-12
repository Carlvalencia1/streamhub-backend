package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type GetStreamByID struct {
	repo domain.Repository
}

func NewGetStreamByID(repo domain.Repository) *GetStreamByID {
	return &GetStreamByID{repo: repo}
}

func (uc *GetStreamByID) Execute(ctx context.Context, streamID string) (*domain.Stream, error) {
	return uc.repo.GetByID(ctx, streamID)
}
