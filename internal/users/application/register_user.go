package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/security"
	"github.com/Carlvalencia1/streamhub-backend/internal/users/domain"
)

type RegisterUser struct {
	repo domain.Repository
}

func NewRegisterUser(repo domain.Repository) *RegisterUser {
	return &RegisterUser{repo: repo}
}

func (uc *RegisterUser) Execute(ctx context.Context, username, email, password string) error {

	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:       uuid.NewString(),
		Username: username,
		Email:    email,
		Password: hash,
	}

	return uc.repo.Create(ctx, user)
}