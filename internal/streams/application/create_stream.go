package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type CreateStream struct {
	repo domain.Repository
}

func NewCreateStream(repo domain.Repository) *CreateStream {
	return &CreateStream{repo: repo}
}

func (uc *CreateStream) Execute(
	ctx context.Context,
	title string,
	description string,
	thumbnail string,
	category string,
	ownerID string,
) error {

	stream := &domain.Stream{
		ID:           uuid.NewString(),
		Title:        title,
		Description:  description,
		ThumbnailURL: thumbnail,
		Category:     category,
		OwnerID:      ownerID,
		IsLive:       false,
		CreatedAt:    time.Now(),
	}

	return uc.repo.Create(ctx, stream)
}