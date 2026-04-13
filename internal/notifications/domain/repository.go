package domain

import "context"

// NotificationRepository define las operaciones de persistencia para notificaciones
type NotificationRepository interface {
	// SaveDeviceToken guarda o actualiza un token de dispositivo (upsert)
	SaveDeviceToken(ctx context.Context, token *DeviceToken) error

	// RemoveDeviceToken elimina un token específico de un usuario
	RemoveDeviceToken(ctx context.Context, userID, token string) error

	// GetDeviceTokensByUser obtiene todos los tokens válidos de un usuario
	GetDeviceTokensByUser(ctx context.Context, userID string) ([]*DeviceToken, error)

	// GetDeviceTokensByUsersExcept obtiene tokens válidos de múltiples usuarios (excepto el ownerID)
	GetDeviceTokensByUsersExcept(ctx context.Context, excludeUserID string) ([]*DeviceToken, error)

	// MarkTokenAsInvalid marca un token como inválido
	MarkTokenAsInvalid(ctx context.Context, token string) error

	// RemoveInvalidTokens elimina todos los tokens marcados como inválidos
	RemoveInvalidTokens(ctx context.Context) error

	// UpdateTokenLastUsed actualiza el timestamp de último uso
	UpdateTokenLastUsed(ctx context.Context, token string) error
}

// PushProvider define la interfaz para enviar notificaciones push
type PushProvider interface {
	// SendMulticast envía una notificación a múltiples dispositivos
	SendMulticast(ctx context.Context, tokens []string, payload *PushPayload) error

	// IsTokenInvalid verifica si un error indica que el token es inválido
	IsTokenInvalid(err error) bool
}

// PushPayload representa el payload de una notificación push
type PushPayload struct {
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]string      `json:"data"`
	AndroidData map[string]interface{} `json:"android_data,omitempty"`
}
