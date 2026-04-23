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
	client            *messaging.Client
	tokenRepository   domain.NotificationRepository
}

func NewFirebasePushProvider(credentialsPath string) (*FirebasePushProvider, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
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

// SetTokenRepository inyecta el repositorio para marcar tokens como inválidos
func (p *FirebasePushProvider) SetTokenRepository(repo domain.NotificationRepository) {
	p.tokenRepository = repo
}

// SendMulticast envía una notificación a múltiples dispositivos
func (p *FirebasePushProvider) SendMulticast(ctx context.Context, tokens []string, payload *domain.PushPayload) error {
	if len(tokens) == 0 {
		logger.Warn("no tokens provided for multicast")
		return nil
	}

	// Convertir data a map[string]string si es necesario
	data := payload.Data
	if data == nil {
		data = make(map[string]string)
	}

	// Construir mensaje
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: payload.Title,
			Body:  payload.Body,
		},
		Data: data,
		// Configuración específica para Android
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title: payload.Title,
				Body:  payload.Body,
				Sound: "default",
			},
			Data: data,
		},
	}

	// Enviar multicast (máximo 500 tokens por llamada)
	resp, err := p.client.SendMulticast(ctx, message)
	if err != nil {
		logger.Error(fmt.Sprintf("error sending multicast: %v", err))
		return err
	}

	// Procesar tokens fallidos y marcarlos como inválidos
	if resp.FailureCount > 0 && p.tokenRepository != nil {
		for idx, err := range resp.Errors {
			if err != nil {
				if idx < len(tokens) {
					failedToken := tokens[idx]
					logger.Warn(fmt.Sprintf("token failed to send: %s, error: %v", failedToken, err.Err))
					
					// Marcar token como inválido en la BD
					if markErr := p.tokenRepository.MarkTokenAsInvalid(ctx, failedToken); markErr != nil {
						logger.Error(fmt.Sprintf("failed to mark token as invalid: %v", markErr))
					}
				}
			}
		}
	}

	// Log de resultados
	logger.Info(fmt.Sprintf("multicast sent successfully. Success: %d, Failure: %d",
		resp.SuccessCount, resp.FailureCount))

	return nil
}

// IsTokenInvalid verifica si un error indica que el token es inválido
func (p *FirebasePushProvider) IsTokenInvalid(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Firebase error codes para tokens inválidos
	invalidTokenErrors := []string{
		"registration token is invalid",
		"invalid registration token provided",
		"mismatched credential",
		"instance id error",
	}

	for _, invalidErr := range invalidTokenErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(invalidErr)) {
			return true
		}
	}

	return false
}

// SendMulticastBatch divide los tokens en lotes y envía
// Útil cuando tenemos más de 500 tokens (límite de Firebase)
func (p *FirebasePushProvider) SendMulticastBatch(ctx context.Context, tokens []string, payload *domain.PushPayload, batchSize int) error {
	if len(tokens) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = 500 // Límite de Firebase
	}

	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}

		batch := tokens[i:end]
		if err := p.SendMulticast(ctx, batch, payload); err != nil {
			logger.Error(fmt.Sprintf("error sending batch [%d:%d]: %v", i, end, err))
			return err
		}
	}

	return nil
}
