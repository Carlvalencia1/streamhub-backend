package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/users/domain"
)

// UserRepo exposes the minimal user repository interface needed by HTTP handlers.
type UserRepo interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateRole(ctx context.Context, userID, role string) error
}
