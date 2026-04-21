package application

import (
	"context"
	"fmt"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

type NotifyNewFollower struct {
	repo     domain.NotificationRepository
	provider domain.PushProvider
}

func NewNotifyNewFollower(repo domain.NotificationRepository, provider domain.PushProvider) *NotifyNewFollower {
	return &NotifyNewFollower{repo: repo, provider: provider}
}

func (uc *NotifyNewFollower) Execute(ctx context.Context, input interface{}) error {
	m, ok := input.(map[string]interface{})
	if !ok {
		return ErrInvalidInput
	}
	streamerID, _ := m["streamer_id"].(string)
	followerUsername, _ := m["follower_username"].(string)

	if streamerID == "" {
		return ErrInvalidInput
	}

	if uc.provider == nil {
		logger.Warn("Firebase provider not initialized, skipping new follower notification")
		return nil
	}

	tokens, err := uc.repo.GetDeviceTokensByUser(ctx, streamerID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get tokens for streamer %s: %v", streamerID, err))
		return err
	}
	if len(tokens) == 0 {
		logger.Info(fmt.Sprintf("no devices for streamer %s, skipping new follower notification", streamerID))
		return nil
	}

	tokenStrings := make([]string, len(tokens))
	for i, t := range tokens {
		tokenStrings[i] = t.Token
	}

	body := "Tienes un nuevo seguidor"
	if followerUsername != "" {
		body = fmt.Sprintf("%s ahora te sigue", followerUsername)
	}

	payload := &domain.PushPayload{
		Title: "Nuevo seguidor",
		Body:  body,
		Data: map[string]string{
			"type":    "new_follower",
			"title":   "Nuevo seguidor",
			"message": body,
		},
	}

	if err := uc.provider.SendMulticast(ctx, tokenStrings, payload); err != nil {
		logger.Error(fmt.Sprintf("failed to send new follower notification: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("new follower notification sent to streamer %s", streamerID))
	return nil
}
