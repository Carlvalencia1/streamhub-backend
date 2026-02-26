package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/chat/domain"
)

type GetMessages struct {
	repo domain.Repository
}

func NewGetMessages(repo domain.Repository) *GetMessages {
	return &GetMessages{repo: repo}
}

func (uc *GetMessages) Execute(ctx context.Context, streamID string) ([]*domain.Message, error) {
	return uc.repo.GetByStream(ctx, streamID)
}