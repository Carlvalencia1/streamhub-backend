package application

import (
	"context"
	"fmt"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

type NotifyStreamLive struct {
	repo     domain.NotificationRepository
	provider domain.PushProvider
}

func NewNotifyStreamLive(repo domain.NotificationRepository, provider domain.PushProvider) *NotifyStreamLive {
	return &NotifyStreamLive{
		repo:     repo,
		provider: provider,
	}
}

type NotifyStreamLiveInput struct {
	StreamID    string
	StreamTitle string
	OwnerUserID string
}

func (uc *NotifyStreamLive) Execute(ctx context.Context, input interface{}) error {
	var notifyInput NotifyStreamLiveInput

	if m, ok := input.(map[string]interface{}); ok {
		if streamID, ok := m["stream_id"].(string); ok {
			notifyInput.StreamID = streamID
		}
		if streamTitle, ok := m["stream_title"].(string); ok {
			notifyInput.StreamTitle = streamTitle
		}
		if ownerID, ok := m["owner_user_id"].(string); ok {
			notifyInput.OwnerUserID = ownerID
		}
	} else if typed, ok := input.(NotifyStreamLiveInput); ok {
		notifyInput = typed
	} else {
		logger.Warn("invalid input type for NotifyStreamLive")
		return ErrInvalidInput
	}

	if notifyInput.StreamID == "" || notifyInput.StreamTitle == "" || notifyInput.OwnerUserID == "" {
		logger.Warn("invalid input for NotifyStreamLive")
		return ErrInvalidInput
	}

	if uc.provider == nil {
		logger.Warn("Firebase provider not initialized, skipping push notifications")
		return nil
	}

	tokens, err := uc.repo.GetDeviceTokensByFollowers(ctx, notifyInput.OwnerUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get device tokens: %v", err))
		return err
	}

	if len(tokens) == 0 {
		logger.Info(fmt.Sprintf("no devices to notify for stream %s", notifyInput.StreamID))
		return nil
	}

	tokenStrings := make([]string, len(tokens))
	for i, t := range tokens {
		tokenStrings[i] = t.Token
	}

	payload := &domain.PushPayload{
		Title: "Stream en vivo",
		Body:  fmt.Sprintf("%s está transmitiendo", notifyInput.StreamTitle),
		Data: map[string]string{
			"type":         "stream_live",
			"stream_id":    notifyInput.StreamID,
			"stream_title": notifyInput.StreamTitle,
			"title":        "Stream en vivo",
			"message":      fmt.Sprintf("%s está en vivo", notifyInput.StreamTitle),
		},
	}

	if err := uc.provider.SendMulticast(ctx, tokenStrings, payload); err != nil {
		logger.Error(fmt.Sprintf("failed to send multicast notification: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("stream live notification sent to %d devices for stream: %s", len(tokenStrings), notifyInput.StreamID))

	for _, token := range tokens {
		_ = uc.repo.UpdateTokenLastUsed(ctx, token.Token)
	}

	return nil
}
