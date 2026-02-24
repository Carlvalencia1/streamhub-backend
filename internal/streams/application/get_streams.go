package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type GetStreams struct {
	repo domain.Repository
}

func NewGetStreams(repo domain.Repository) *GetStreams {
	return &GetStreams{repo: repo}
}

func (uc *GetStreams) Execute(ctx context.Context) ([]*domain.Stream, error) {
	return uc.repo.GetAll(ctx)
}