package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/Carlvalencia1/streamhub-backend/internal/chat/domain"
)

type SendMessage struct {
	repo domain.Repository
}

func NewSendMessage(repo domain.Repository) *SendMessage {
	return &SendMessage{repo: repo}
}

func (uc *SendMessage) Execute(
	ctx context.Context,
	streamID string,
	userID string,
	content string,
) (*domain.Message, error) {

	msg := &domain.Message{
		ID:        uuid.NewString(),
		StreamID:  streamID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	err := uc.repo.Save(ctx, msg)
	return msg, err
}