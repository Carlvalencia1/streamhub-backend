package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

type FirebasePushProvider struct {
	client          *messaging.Client
	tokenRepository domain.NotificationRepository
}

func NewFirebasePushProvider(credentialsPath string) (*FirebasePushProvider, error) {
	ctx := context.Background()

	if credentialsPath == "" {
		logger.Error("CRITICAL: credentialsPath is empty")
		return nil, fmt.Errorf("firebase credentials path is required")
	}

	logger.Info(fmt.Sprintf("Initializing Firebase with credentials: %s", credentialsPath))

	opt := option.WithCredentialsFile(credentialsPath)
	opt = option.WithEndpoint("https://fcm.googleapis.com/v1")
	
	conf := &firebase.Config{
		ProjectID: "streamhub-64704",
	}

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		logger.Error(fmt.Sprintf("error initializing Firebase app: %v", err))
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("error creating Firebase Messaging client: %v", err))
		return nil, err
	}

	logger.Info("Firebase Messaging client initialized successfully")
	return &FirebasePushProvider{client: client}, nil
}

func (p *FirebasePushProvider) SetTokenRepository(repo domain.NotificationRepository) {
	p.tokenRepository = repo
}

// SendMulticast envía notificaciones a múltiples dispositivos usando SendEachForMulticast
func (p *FirebasePushProvider) SendMulticast(ctx context.Context, tokens []string, payload *domain.PushPayload) error {
	if len(tokens) == 0 {
		logger.Warn("no tokens provided for multicast")
		return nil
	}

	logger.Info(fmt.Sprintf("Sending multicast to %d tokens", len(tokens)))

	data := payload.Data
	if data == nil {
		data = make(map[string]string)
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: payload.Title,
			Body:  payload.Body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title: payload.Title,
				Body:  payload.Body,
				Sound: "default",
			},
		},
	}

	traceID := ""
	if payload.Data != nil {
		if tid, ok := payload.Data["trace_id"]; ok {
			traceID = tid
		}
	}

	// 🔥 CAMBIO IMPORTANTE: usar SendEachForMulticast en lugar de SendMulticast
	resp, err := p.client.SendEachForMulticast(ctx, message)
	if err != nil {
		logger.Error(fmt.Sprintf("[%s] error sending multicast: %v", traceID, err))
		return err
	}

	invalidatedCount := 0
	if resp.FailureCount > 0 && p.tokenRepository != nil {
		for idx, sendResp := range resp.Responses {
			if sendResp.Error != nil && idx < len(tokens) {
				failedToken := tokens[idx]
				logger.Warn(fmt.Sprintf("[%s] token failed: %s, error: %v", traceID, failedToken, sendResp.Error))
				if err := p.tokenRepository.MarkTokenAsInvalid(ctx, failedToken); err == nil {
					invalidatedCount++
				}
			}
		}
	}

	logger.Info(fmt.Sprintf("[%s] multicast result - Success: %d, Failure: %d, Invalidated: %d",
		traceID, resp.SuccessCount, resp.FailureCount, invalidatedCount))

	return nil
}

func (p *FirebasePushProvider) IsTokenInvalid(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	invalidTokenErrors := []string{
		"registration token is invalid",
		"invalid registration token",
		"mismatched credential",
		"unregistered",
		"not found",
	}
	for _, invalidErr := range invalidTokenErrors {
		if strings.Contains(strings.ToLower(errStr), invalidErr) {
			return true
		}
	}
	return false
}

func (p *FirebasePushProvider) SendMulticastBatch(ctx context.Context, tokens []string, payload *domain.PushPayload, traceID string) error {
	if len(tokens) == 0 {
		return nil
	}

	batchSize := 500
	totalBatches := (len(tokens) + batchSize - 1) / batchSize

	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]
		batchNum := (i / batchSize) + 1

		logger.Info(fmt.Sprintf("[%s] sending batch %d/%d with %d tokens", traceID, batchNum, totalBatches, len(batch)))

		if err := p.SendMulticast(ctx, batch, payload); err != nil {
			logger.Error(fmt.Sprintf("[%s] error sending batch %d: %v", traceID, batchNum, err))
			return err
		}
	}
	return nil
}