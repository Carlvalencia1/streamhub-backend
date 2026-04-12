package domain

import "context"

type Repository interface {
	Create(ctx context.Context, stream *Stream) error
	GetAll(ctx context.Context) ([]*Stream, error)
	GetByID(ctx context.Context, streamID string) (*Stream, error)
	StartStream(ctx context.Context, streamID string) error
	StopStream(ctx context.Context, streamID string) error
	JoinStream(ctx context.Context, streamID string) error
}