package domain

import "context"

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateRole(ctx context.Context, userID, role string) error
	List(ctx context.Context) ([]*User, error)
}
