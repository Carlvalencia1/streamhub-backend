package application

import (
	"context"

	"github.com/Carlvalencia1/streamhub-backend/internal/users/domain"
)

type ListUsers struct {
	repo domain.Repository
}

func NewListUsers(repo domain.Repository) *ListUsers {
	return &ListUsers{repo: repo}
}

func (uc *ListUsers) Execute(ctx context.Context) ([]*domain.User, error) {
	return uc.repo.List(ctx)
}