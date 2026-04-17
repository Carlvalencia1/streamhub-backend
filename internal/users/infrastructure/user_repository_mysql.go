package infrastructure

import (
	"context"
	"database/sql"

	"github.com/Carlvalencia1/streamhub-backend/internal/users/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
INSERT INTO users (id, username, email, password, role, avatar_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.AvatarURL,
	)
	return err
}

func (r *MySQLRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
SELECT id, username, email, password, role, nickname, bio, location, avatar_url, created_at, updated_at
FROM users WHERE email = ?
`
	row := r.db.QueryRowContext(ctx, query, email)
	var user domain.User
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Role,
		&user.Nickname, &user.Bio, &user.Location, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MySQLRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
SELECT id, username, email, password, role, nickname, bio, location, avatar_url, created_at, updated_at
FROM users WHERE id = ?
`
	row := r.db.QueryRowContext(ctx, query, id)
	var user domain.User
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Role,
		&user.Nickname, &user.Bio, &user.Location, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MySQLRepository) UpdateRole(ctx context.Context, userID, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = ?, updated_at = NOW() WHERE id = ?`, role, userID)
	return err
}

func (r *MySQLRepository) UpdateProfile(ctx context.Context, userID string, nickname, bio, location *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET nickname = ?, bio = ?, location = ?, updated_at = NOW() WHERE id = ?`,
		nickname, bio, location, userID,
	)
	return err
}

func (r *MySQLRepository) List(ctx context.Context) ([]*domain.User, error) {
	query := `
SELECT id, username, email, password, role, nickname, bio, location, avatar_url, created_at, updated_at
FROM users
`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.Role,
			&user.Nickname, &user.Bio, &user.Location, &user.AvatarURL,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}
