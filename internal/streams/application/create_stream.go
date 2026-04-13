package application

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

// Get RTMP base URL from environment or use default
// Default: rtmp://54.144.66.251/live (Streaming server)
func getRTMPBaseURL() string {
	if url := os.Getenv("RTMP_BASE_URL"); url != "" {
		return url
	}
	return "rtmp://54.144.66.251/live"
}

// Get HLS base URL from environment or use default
// Default: http://54.144.66.251/live (Streaming server)
func getHLSBaseURL() string {
	if url := os.Getenv("HLS_BASE_URL"); url != "" {
		return url
	}
	return "http://54.144.66.251/live"
}

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
	hlsBaseURL := getHLSBaseURL()
	playbackURL := fmt.Sprintf("%s/%s/index.m3u8", hlsBaseURL, streamKey)

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