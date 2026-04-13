package application

import "context"

// StreamLiveNotifier interfaz para notificar cuando un stream va en vivo
// Evita importaciones cíclicas entre módulos
type StreamLiveNotifier interface {
	Execute(ctx context.Context, input interface{}) error
}

// Variable global para inyectar dependencia de notificaciones
var streamLiveNotifier StreamLiveNotifier

// SetStreamLiveNotifier inyecta el notificador de stream en vivo
// Llamado desde router.go después de inicializar Firebase
func SetStreamLiveNotifier(notifier StreamLiveNotifier) {
	streamLiveNotifier = notifier
}

// GetStreamLiveNotifier obtiene el notificador si fue inyectado
func GetStreamLiveNotifier() StreamLiveNotifier {
	return streamLiveNotifier
}

