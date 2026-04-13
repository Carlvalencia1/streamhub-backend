package domain

import "time"

// DeviceToken representa un token de dispositivo registrado para FCM
type DeviceToken struct {
	ID         string
	UserID     string
	Token      string
	Platform   string    // "android", "ios"
	DeviceID   string    // ID único del dispositivo
	AppVersion string    // Versión de la app
	IsValid    bool      // Token válido (no unregistered/invalid)
	LastUsedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewDeviceToken crea una nueva instancia de DeviceToken
func NewDeviceToken(id, userID, token, platform, deviceID, appVersion string) *DeviceToken {
	now := time.Now()
	return &DeviceToken{
		ID:         id,
		UserID:     userID,
		Token:      token,
		Platform:   platform,
		DeviceID:   deviceID,
		AppVersion: appVersion,
		IsValid:    true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
