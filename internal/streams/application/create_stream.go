package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

const (
	RTMPServerURL    = "rtmp://3.232.197.126/live"
	HLSServerURL     = "http://3.232.197.126/live"
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
) (*domain.Stream, error) {

	streamKey := uuid.NewString()
	rtmpURL := fmt.Sprintf("%s/%s", RTMPServerURL, streamKey)
	playbackURL := fmt.Sprintf("%s/%s.m3u8", HLSServerURL, streamKey)

	stream := &domain.Stream{
		ID:           uuid.NewString(),
		Title:        title,
		Description:  description,
		ThumbnailURL: thumbnail,
		Category:     category,
		OwnerID:      ownerID,
		IsLive:       false,
		CreatedAt:    time.Now(),
		StreamKey:    streamKey,
		PlaybackURL:  playbackURL,
	}

	err := uc.repo.Create(ctx, stream)
	if err != nil {
		return nil, err
	}

	return stream, nil
}