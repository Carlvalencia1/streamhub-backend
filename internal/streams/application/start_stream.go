package application

import (
	"context"
	"fmt"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type StartStream struct {
	repo domain.Repository
}

func NewStartStream(repo domain.Repository) *StartStream {
	return &StartStream{repo: repo}
}

// StreamRepositoryExt es una interfaz extendida que combina Repository basico y StreamRepository
type StreamRepositoryExt interface {
	domain.Repository
	domain.StreamRepository
}

func (uc *StartStream) Execute(ctx context.Context, streamID string) error {
	logger.Debug(fmt.Sprintf("StartStream usecase started for stream: %s", streamID))

	// 1. Iniciar stream en la BD
	if err := uc.repo.StartStream(ctx, streamID); err != nil {
		logger.Error(fmt.Sprintf("failed to start stream: %v", err))
		return err
	}

	// 2. Obtener datos del stream para notificaciones
	streamRepoExt, ok := uc.repo.(StreamRepositoryExt)
	if !ok {
		logger.Warn("repository does not implement GetByID, skipping notifications")
		return nil
	}

	stream, err := streamRepoExt.GetByID(ctx, streamID)
	if err != nil {
		logger.Warn(fmt.Sprintf("could not fetch stream for notifications: %v", err))
		return nil // No fallar si no podemos notificar
	}

	// 3. Enviar notificación si está disponible el notificador
	notifier := GetStreamLiveNotifier()
	if notifier != nil {
		// Usar la interfaz genérica con un tipo que contenga los datos necesarios
		notifyInput := map[string]interface{}{
			"stream_id":     stream.ID,
			"stream_title":  stream.Title,
			"owner_user_id": stream.OwnerID,
		}

		go func() {
			// Ejecutar notificación en background para no bloquear
			if err := notifier.Execute(context.Background(), notifyInput); err != nil {
				logger.Error(fmt.Sprintf("failed to send stream live notification: %v", err))
			}
		}()
	} else {
		logger.Warn("stream live notifier not initialized")
	}

	logger.Info(fmt.Sprintf("stream %s started successfully", streamID))
	return nil
}