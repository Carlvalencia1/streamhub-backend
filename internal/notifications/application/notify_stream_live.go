package application

import (
	"context"
	"fmt"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	usersDomain "github.com/Carlvalencia1/streamhub-backend/internal/users/domain"
	"github.com/google/uuid"
)

type NotifyStreamLive struct {
	repo     domain.NotificationRepository
	provider domain.PushProvider
	userRepo usersDomain.Repository
}

func NewNotifyStreamLive(repo domain.NotificationRepository, provider domain.PushProvider, userRepo usersDomain.Repository) *NotifyStreamLive {
	return &NotifyStreamLive{
		repo:     repo,
		provider: provider,
		userRepo: userRepo,
	}
}

type NotifyStreamLiveInput struct {
	StreamID    string
	StreamTitle string
	OwnerUserID string
}

func (uc *NotifyStreamLive) Execute(ctx context.Context, input interface{}) error {
	// Generar trace_id único para E2E tracking
	traceID := uuid.NewString()

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

	// Log inicio de solicitud
	logger.NotificationStreamLiveRequest(traceID, notifyInput.StreamID, notifyInput.OwnerUserID)

	// Obtener tokens de followers
	followerTokens, err := uc.repo.GetDeviceTokensByFollowers(ctx, notifyInput.OwnerUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("[%s] failed to get follower tokens: %v", traceID, err))
		return err
	}

	// Obtener tokens del streamer para autonotificación
	streamerTokens, err := uc.repo.GetDeviceTokensByUser(ctx, notifyInput.OwnerUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("[%s] failed to get streamer tokens: %v", traceID, err))
		return err
	}

	// Deduplicar tokens (usar map para eliminar duplicados)
	tokenSet := make(map[string]bool)
	var allTokens []string

	// Agregar tokens de followers
	for _, token := range followerTokens {
		if !tokenSet[token.Token] {
			tokenSet[token.Token] = true
			allTokens = append(allTokens, token.Token)
		}
	}

	// Agregar tokens del streamer (autonotificación)
	for _, token := range streamerTokens {
		if !tokenSet[token.Token] {
			tokenSet[token.Token] = true
			allTokens = append(allTokens, token.Token)
		}
	}

	followerTokenCount := len(followerTokens)
	streamerTokenCount := len(streamerTokens)
	dedupTokenCount := len(allTokens)

	// Log de resolución de destinatarios
	logger.ResolveRecipients(traceID, followerTokenCount, streamerTokenCount, dedupTokenCount)

	if dedupTokenCount == 0 {
		logger.Info(fmt.Sprintf("[%s] no devices to notify for stream %s", traceID, notifyInput.StreamID))
		return nil
	}

	// Obtener nombre del streamer
	streamerName := notifyInput.StreamTitle
	if uc.userRepo != nil {
		if u, err := uc.userRepo.GetByID(ctx, notifyInput.OwnerUserID); err == nil && u != nil {
			streamerName = u.Username
		}
	}

	// Construir payload con trace_id y streamer_id
	payload := &domain.PushPayload{
		Title: "Stream en vivo",
		Body:  fmt.Sprintf("%s está en vivo", streamerName),
		Data: map[string]string{
			"type":         "stream_live",
			"stream_id":    notifyInput.StreamID,
			"stream_title": notifyInput.StreamTitle,
			"streamer_id":  notifyInput.OwnerUserID,
			"title":        "Stream en vivo",
			"message":      fmt.Sprintf("%s está en vivo", streamerName),
			"trace_id":     traceID,
		},
	}

	// Enviar con batching (máximo 500 tokens por lote)
	if err := uc.provider.SendMulticastBatch(ctx, allTokens, payload, traceID); err != nil {
		logger.Error(fmt.Sprintf("[%s] failed to send multicast notification: %v", traceID, err))
		return err
	}

	// Actualizar last_used_at para todos los tokens enviados
	for _, token := range followerTokens {
		_ = uc.repo.UpdateTokenLastUsed(ctx, token.Token)
	}
	for _, token := range streamerTokens {
		_ = uc.repo.UpdateTokenLastUsed(ctx, token.Token)
	}

	// Log de finalización
	logger.StreamLiveNotificationDone(traceID, dedupTokenCount)

	return nil
}
