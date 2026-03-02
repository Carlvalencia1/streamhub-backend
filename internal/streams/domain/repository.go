package domain

import "context"

type Repository interface {
	Create(ctx context.Context, stream *Stream) error
	GetAll(ctx context.Context) ([]*Stream, error)
	StartStream(ctx context.Context, streamID string) error
	JoinStream(ctx context.Context, streamID string) error
}