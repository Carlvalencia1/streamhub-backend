package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

type RemoveFcmToken struct {
	repo domain.NotificationRepository
}

func NewRemoveFcmToken(repo domain.NotificationRepository) *RemoveFcmToken {
	return &RemoveFcmToken{repo: repo}
}

type RemoveFcmTokenInput struct {
	UserID string
	Token  string
}

// Execute elimina un token FCM específico del usuario
func (uc *RemoveFcmToken) Execute(ctx context.Context, input RemoveFcmTokenInput) error {
	logger.Debug("RemoveFcmToken usecase started for user: " + input.UserID)

	// Validación
	if input.UserID == "" || input.Token == "" {
		logger.Warn("invalid input: user_id or token missing")
		return ErrInvalidInput
	}

	// Eliminar
	if err := uc.repo.RemoveDeviceToken(ctx, input.UserID, input.Token); err != nil {
		logger.Error("failed to remove device token: " + err.Error())
		return err
	}

	logger.Info("FCM token removed for user: " + input.UserID)
	return nil
}
