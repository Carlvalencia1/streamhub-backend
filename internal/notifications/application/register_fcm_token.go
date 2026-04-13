package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/google/uuid"
)

type RegisterFcmToken struct {
	repo domain.NotificationRepository
}

func NewRegisterFcmToken(repo domain.NotificationRepository) *RegisterFcmToken {
	return &RegisterFcmToken{repo: repo}
}

type RegisterFcmTokenInput struct {
	UserID     string
	Token      string
	Platform   string // "android" o "ios"
	DeviceID   string // opcional
	AppVersion string // opcional
}

// Execute registra un nuevo token FCM para el usuario (upsert)
func (uc *RegisterFcmToken) Execute(ctx context.Context, input RegisterFcmTokenInput) error {
	logger.Debug("RegisterFcmToken usecase started for user: " + input.UserID)

	// Validación básica
	if input.UserID == "" || input.Token == "" {
		logger.Warn("invalid input: user_id or token missing")
		return ErrInvalidInput
	}

	// Si no se proporciona platform, usar android por defecto
	if input.Platform == "" {
		input.Platform = "android"
	}

	// Crear entity
	deviceToken := domain.NewDeviceToken(
		uuid.NewString(),
		input.UserID,
		input.Token,
		input.Platform,
		input.DeviceID,
		input.AppVersion,
	)

	// Guardar
	if err := uc.repo.SaveDeviceToken(ctx, deviceToken); err != nil {
		logger.Error("failed to save device token: " + err.Error())
		return err
	}

	logger.Info("FCM token registered successfully for user: " + input.UserID)
	return nil
}
