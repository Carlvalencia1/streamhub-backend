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
INSERT INTO users (id, username, email, password, avatar_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW(), NOW())
`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.AvatarURL,
	)

	return err
}

func (r *MySQLRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {

	query := `
	SELECT id, username, email, password, avatar_url, created_at, updated_at
	FROM users
	WHERE email = ?
	`

	row := r.db.QueryRowContext(ctx, query, email)

	var user domain.User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *MySQLRepository) List(ctx context.Context) ([]*domain.User, error) {

	query := `
	SELECT id, username, email, password, avatar_url, created_at, updated_at
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
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.AvatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}