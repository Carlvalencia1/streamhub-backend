package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/google/uuid"
)

type DeviceTokenRepository struct {
	db *sql.DB
}

func NewDeviceTokenRepository(db *sql.DB) *DeviceTokenRepository {
	return &DeviceTokenRepository{db: db}
}

// SaveDeviceToken guarda o actualiza un token (upsert)
func (r *DeviceTokenRepository) SaveDeviceToken(ctx context.Context, token *domain.DeviceToken) error {
	if token.ID == "" {
		token.ID = uuid.NewString()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}
	token.UpdatedAt = time.Now()

	query := `
		INSERT INTO device_tokens (id, user_id, token, platform, device_id, app_version, is_valid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			platform = VALUES(platform),
			device_id = VALUES(device_id),
			app_version = VALUES(app_version),
			is_valid = true,
			updated_at = VALUES(updated_at)
	`

	result, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.Token,
		token.Platform,
		token.DeviceID,
		token.AppVersion,
		token.IsValid,
		token.CreatedAt,
		token.UpdatedAt,
	)

	if err != nil {
		logger.Error(fmt.Sprintf("failed to save device token for user %s: %v", token.UserID, err))
		return err
	}

	rows, _ := result.RowsAffected()
	logger.Debug(fmt.Sprintf("device token operation completed, rows affected: %d", rows))

	return nil
}

// RemoveDeviceToken elimina un token específico de un usuario
func (r *DeviceTokenRepository) RemoveDeviceToken(ctx context.Context, userID, token string) error {
	query := `DELETE FROM device_tokens WHERE user_id = ? AND token = ?`

	result, err := r.db.ExecContext(ctx, query, userID, token)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to remove device token for user %s: %v", userID, err))
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		logger.Warn(fmt.Sprintf("token not found for user %s", userID))
	}

	return nil
}

// GetDeviceTokensByUser obtiene todos los tokens válidos de un usuario
func (r *DeviceTokenRepository) GetDeviceTokensByUser(ctx context.Context, userID string) ([]*domain.DeviceToken, error) {
	query := `
		SELECT id, user_id, token, platform, device_id, app_version, is_valid, last_used_at, created_at, updated_at
		FROM device_tokens
		WHERE user_id = ? AND is_valid = true
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get device tokens for user %s: %v", userID, err))
		return nil, err
	}
	defer rows.Close()

	return r.scanDeviceTokens(rows)
}

// GetDeviceTokensByUsersExcept obtiene tokens válidos de todos los usuarios excepto uno
// Usado para notificar a otros usuarios cuando alguien inicia stream
func (r *DeviceTokenRepository) GetDeviceTokensByUsersExcept(ctx context.Context, excludeUserID string) ([]*domain.DeviceToken, error) {
	query := `
		SELECT id, user_id, token, platform, device_id, app_version, is_valid, last_used_at, created_at, updated_at
		FROM device_tokens
		WHERE user_id != ? AND is_valid = true
		ORDER BY user_id, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, excludeUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get device tokens excluding user %s: %v", excludeUserID, err))
		return nil, err
	}
	defer rows.Close()

	return r.scanDeviceTokens(rows)
}

// MarkTokenAsInvalid marca un token como inválido
// Se usa cuando Firebase reporta que el token no es válido
func (r *DeviceTokenRepository) MarkTokenAsInvalid(ctx context.Context, token string) error {
	query := `UPDATE device_tokens SET is_valid = false, updated_at = NOW() WHERE token = ?`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to mark token as invalid: %v", err))
		return err
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		logger.Debug(fmt.Sprintf("marked token as invalid"))
	}

	return nil
}

// RemoveInvalidTokens elimina todos los tokens marcados como inválidos
func (r *DeviceTokenRepository) RemoveInvalidTokens(ctx context.Context) error {
	query := `DELETE FROM device_tokens WHERE is_valid = false`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to remove invalid tokens: %v", err))
		return err
	}

	rows, _ := result.RowsAffected()
	logger.Info(fmt.Sprintf("removed %d invalid device tokens", rows))

	return nil
}

// UpdateTokenLastUsed actualiza el timestamp de último uso
func (r *DeviceTokenRepository) UpdateTokenLastUsed(ctx context.Context, token string) error {
	query := `UPDATE device_tokens SET last_used_at = NOW() WHERE token = ?`

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to update token last used: %v", err))
		return err
	}

	return nil
}

// scanDeviceTokens helper para mapear resultados SQL a DeviceToken
func (r *DeviceTokenRepository) scanDeviceTokens(rows *sql.Rows) ([]*domain.DeviceToken, error) {
	var tokens []*domain.DeviceToken

	for rows.Next() {
		var dt domain.DeviceToken
		err := rows.Scan(
			&dt.ID,
			&dt.UserID,
			&dt.Token,
			&dt.Platform,
			&dt.DeviceID,
			&dt.AppVersion,
			&dt.IsValid,
			&dt.LastUsedAt,
			&dt.CreatedAt,
			&dt.UpdatedAt,
		)
		if err != nil {
			logger.Error(fmt.Sprintf("error scanning device token row: %v", err))
			return nil, err
		}

		tokens = append(tokens, &dt)
	}

	if err := rows.Err(); err != nil {
		logger.Error(fmt.Sprintf("error iterating device token rows: %v", err))
		return nil, err
	}

	return tokens, nil
}
