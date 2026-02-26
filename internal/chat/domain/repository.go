package domain

import "context"

type Repository interface {
	Save(ctx context.Context, msg *Message) error
	GetByStream(ctx context.Context, streamID string) ([]*Message, error)
}