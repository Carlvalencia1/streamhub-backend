package application

import (
	"context"
	"errors"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/security"
	"github.com/Carlvalencia1/streamhub-backend/internal/users/domain"
)

type LoginUser struct {
	repo domain.Repository
}

func NewLoginUser(repo domain.Repository) *LoginUser {
	return &LoginUser{repo: repo}
}

func (uc *LoginUser) Execute(ctx context.Context, email, password string) (string, error) {

	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := security.CheckPassword(user.Password, password); err != nil {
		return "", errors.New("invalid credentials")
	}

	return security.GenerateToken(user.ID, user.Username)
}