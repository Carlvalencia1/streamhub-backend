package domain

import "time"

type DeviceToken struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Token      string     `json:"token"`
	Platform   string     `json:"platform"`
	DeviceID   string     `json:"device_id"`
	AppVersion string     `json:"app_version"`
	IsValid    bool       `json:"is_valid"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

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
