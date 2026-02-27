package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, stream *domain.Stream) error {

	query := `
	INSERT INTO streams 
	(id, title, description, thumbnail_url, category, owner_id, is_live, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		stream.ID,
		stream.Title,
		stream.Description,
		stream.ThumbnailURL,
		stream.Category,
		stream.OwnerID,
		stream.IsLive,
		stream.CreatedAt,
	)

	return err
}

func (r *MySQLRepository) GetAll(ctx context.Context) ([]*domain.Stream, error) {

	query := `
	SELECT id, title, description, thumbnail_url, category,
	       owner_id, viewers_count, is_live, started_at, created_at
	FROM streams
	ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var streams []*domain.Stream

	for rows.Next() {

		var s domain.Stream
		var startedAt sql.NullTime

		err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Description,
			&s.ThumbnailURL,
			&s.Category,
			&s.OwnerID,
			&s.ViewersCount,
			&s.IsLive,
			&startedAt,
			&s.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		if startedAt.Valid {
			t := startedAt.Time
			s.StartedAt = &t
		}

		streams = append(streams, &s)
	}

	return streams, nil
}

func (r *MySQLRepository) JoinStream(ctx context.Context, streamID string) error {

	query := `
	UPDATE streams
	SET viewers_count = viewers_count + 1
	WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, streamID)
	return err
}

func (r *MySQLRepository) StartStream(ctx context.Context, streamID string) error {

	now := time.Now()

	query := `
	UPDATE streams
	SET is_live = true,
	    started_at = ?
	WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, now, streamID)

	return err
}