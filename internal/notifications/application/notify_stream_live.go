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

// NotifyStreamLiveInput estructura esperada en Execute
type NotifyStreamLiveInput struct {
	StreamID    string
	StreamTitle string
	OwnerUserID string
}

// Execute envía notificación push a todos los usuarios excepto el owner
// cuando un stream inicia (is_live = true)
// Acepta interface{} para compatibilidad con StreamLiveNotifier
func (uc *NotifyStreamLive) Execute(ctx context.Context, input interface{}) error {
	// Convertir input genérico a NotifyStreamLiveInput
	var notifyInput NotifyStreamLiveInput

	// Si es map, extraer valores
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
		// Si ya es del tipo correcto
		notifyInput = typed
	} else {
		logger.Warn("invalid input type for NotifyStreamLive")
		return ErrInvalidInput
	}

	logger.Debug(fmt.Sprintf("NotifyStreamLive usecase started for stream: %s", notifyInput.StreamID))

	// Validación
	if notifyInput.StreamID == "" || notifyInput.StreamTitle == "" || notifyInput.OwnerUserID == "" {
		logger.Warn("invalid input for NotifyStreamLive")
		return ErrInvalidInput
	}

	// Si no hay provider de Firebase (no configurado), retornar sin error
	if uc.provider == nil {
		logger.Warn("Firebase provider not initialized, skipping push notifications")
		return nil
	}

	// 1. Obtener todos los tokens válidos excepto del owner
	tokens, err := uc.repo.GetDeviceTokensByUsersExcept(ctx, notifyInput.OwnerUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get device tokens: %v", err))
		return err
	}

	if len(tokens) == 0 {
		logger.Info(fmt.Sprintf("no devices to notify for stream %s", notifyInput.StreamID))
		return nil
	}

	// 2. Construir payload (debe coincidir exacto con Android)
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

	// 3. Enviar multicast
	if err := uc.provider.SendMulticast(ctx, tokenStrings, payload); err != nil {
		logger.Error(fmt.Sprintf("failed to send multicast notification: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("stream live notification sent to %d devices for stream: %s", len(tokenStrings), notifyInput.StreamID))

	// 4. Actualizar last_used_at para los tokens (best effort)
	for _, token := range tokens {
		_ = uc.repo.UpdateTokenLastUsed(ctx, token.Token)
	}

	return nil
}
