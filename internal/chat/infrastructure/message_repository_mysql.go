package infrastructure

import (
	"context"
	"database/sql"

	"github.com/Carlvalencia1/streamhub-backend/internal/chat/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Save(ctx context.Context, msg *domain.Message) error {

	query := `
	INSERT INTO messages (id, stream_id, user_id, content, created_at)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.StreamID,
		msg.UserID,
		msg.Content,
		msg.CreatedAt,
	)

	return err
}

func (r *MySQLRepository) GetByStream(ctx context.Context, streamID string) ([]*domain.Message, error) {

	query := `
	SELECT id, stream_id, user_id, content, created_at
	FROM messages
	WHERE stream_id = ?
	ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, streamID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []*domain.Message

	for rows.Next() {

		var m domain.Message

		err := rows.Scan(
			&m.ID,
			&m.StreamID,
			&m.UserID,
			&m.Content,
			&m.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		messages = append(messages, &m)
	}

	return messages, nil
}