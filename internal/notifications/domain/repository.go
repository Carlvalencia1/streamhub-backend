package domain

import "context"

type NotificationRepository interface {
	SaveDeviceToken(ctx context.Context, token *DeviceToken) error
	RemoveDeviceToken(ctx context.Context, userID, token string) error
	GetDeviceTokensByUser(ctx context.Context, userID string) ([]*DeviceToken, error)
	GetDeviceTokensByUsersExcept(ctx context.Context, excludeUserID string) ([]*DeviceToken, error)
	GetDeviceTokensByFollowers(ctx context.Context, streamerID string) ([]*DeviceToken, error)
	MarkTokenAsInvalid(ctx context.Context, token string) error
	RemoveInvalidTokens(ctx context.Context) error
	UpdateTokenLastUsed(ctx context.Context, token string) error
}

type PushProvider interface {
	SendMulticast(ctx context.Context, tokens []string, payload *PushPayload) error
	IsTokenInvalid(err error) bool
}

type PushPayload struct {
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]string      `json:"data"`
	AndroidData map[string]interface{} `json:"android_data,omitempty"`
}
