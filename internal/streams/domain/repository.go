package domain

import "context"

type StreamRepository interface {
	Create(ctx context.Context, stream *Stream) error
	GetAll(ctx context.Context) ([]*Stream, error)
	GetByID(ctx context.Context, streamID string) (*Stream, error)
	GetByStreamKey(ctx context.Context, streamKey string) (*Stream, error)
	Update(ctx context.Context, stream *Stream) error
	StartStream(ctx context.Context, streamID string) error
	StopStream(ctx context.Context, streamID string) error
	JoinStream(ctx context.Context, streamID string) error
}

// Deprecated: Use StreamRepository instead
type Repository interface {
	Create(ctx context.Context, stream *Stream) error
	GetAll(ctx context.Context) ([]*Stream, error)
	GetByID(ctx context.Context, streamID string) (*Stream, error)
	StartStream(ctx context.Context, streamID string) error
	StopStream(ctx context.Context, streamID string) error
	JoinStream(ctx context.Context, streamID string) error
}