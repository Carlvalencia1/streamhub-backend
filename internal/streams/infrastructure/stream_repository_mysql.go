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
	(id, title, description, thumbnail_url, category, owner_id, is_live, created_at, stream_key, playback_url)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		stream.ID, stream.Title, stream.Description, stream.ThumbnailURL,
		stream.Category, stream.OwnerID, stream.IsLive, stream.CreatedAt,
		stream.StreamKey, stream.PlaybackURL,
	)

	return err
}

func (r *MySQLRepository) GetAll(ctx context.Context) ([]*domain.Stream, error) {

	query := `
	SELECT id, title, description, thumbnail_url, category,
	       owner_id, viewers_count, is_live, started_at, ended_at, created_at, stream_key, playback_url, recording_url
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
		var endedAt sql.NullTime
		var recordingURL sql.NullString

		err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.ThumbnailURL, &s.Category,
			&s.OwnerID, &s.ViewersCount, &s.IsLive,
			&startedAt, &endedAt, &s.CreatedAt, &s.StreamKey, &s.PlaybackURL,
			&recordingURL,
		)
		if err != nil {
			return nil, err
		}
		if startedAt.Valid {
			t := startedAt.Time
			s.StartedAt = &t
		}
		if endedAt.Valid {
			t := endedAt.Time
			s.EndedAt = &t
		}
		if recordingURL.Valid {
			s.RecordingURL = &recordingURL.String
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

func (r *MySQLRepository) StopStream(ctx context.Context, streamID string) error {

	now := time.Now()

	query := `
	UPDATE streams
	SET is_live = false, ended_at = ?
	WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, now, streamID)

	return err
}

func (r *MySQLRepository) GetByID(ctx context.Context, streamID string) (*domain.Stream, error) {

	query := `
	SELECT id, title, description, thumbnail_url, category,
	       owner_id, viewers_count, is_live, started_at, ended_at, created_at, stream_key, playback_url, recording_url
	FROM streams
	WHERE id = ?
	`

	var s domain.Stream
	var startedAt sql.NullTime
	var endedAt sql.NullTime
	var recordingURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, streamID).Scan(
		&s.ID, &s.Title, &s.Description, &s.ThumbnailURL, &s.Category,
		&s.OwnerID, &s.ViewersCount, &s.IsLive,
		&startedAt, &endedAt, &s.CreatedAt, &s.StreamKey, &s.PlaybackURL,
		&recordingURL,
	)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		t := startedAt.Time
		s.StartedAt = &t
	}
	if endedAt.Valid {
		t := endedAt.Time
		s.EndedAt = &t
	}
	if recordingURL.Valid {
		s.RecordingURL = &recordingURL.String
	}
	return &s, nil
}

// GetByStreamKey retrieves a stream by its stream_key
// Used by NGINX RTMP validation
func (r *MySQLRepository) GetByStreamKey(ctx context.Context, streamKey string) (*domain.Stream, error) {

	query := `
	SELECT id, title, description, thumbnail_url, category,
	       owner_id, viewers_count, is_live, started_at, ended_at, created_at, stream_key, playback_url, recording_url
	FROM streams
	WHERE stream_key = ?
	`

	var s domain.Stream
	var startedAt sql.NullTime
	var endedAt sql.NullTime
	var recordingURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, streamKey).Scan(
		&s.ID, &s.Title, &s.Description, &s.ThumbnailURL, &s.Category,
		&s.OwnerID, &s.ViewersCount, &s.IsLive,
		&startedAt, &endedAt, &s.CreatedAt, &s.StreamKey, &s.PlaybackURL,
		&recordingURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if startedAt.Valid {
		t := startedAt.Time
		s.StartedAt = &t
	}
	if endedAt.Valid {
		t := endedAt.Time
		s.EndedAt = &t
	}
	if recordingURL.Valid {
		s.RecordingURL = &recordingURL.String
	}
	return &s, nil
}

func (r *MySQLRepository) SetRecordingURL(ctx context.Context, streamKey, recordingURL string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE streams SET recording_url = ? WHERE stream_key = ?`,
		recordingURL, streamKey,
	)
	return err
}

// Update updates an existing stream
// Used by NGINX RTMP validation to mark stream as active/inactive
func (r *MySQLRepository) Update(ctx context.Context, stream *domain.Stream) error {

	query := `
	UPDATE streams
	SET title = ?,
	    description = ?,
	    thumbnail_url = ?,
	    category = ?,
	    owner_id = ?,
	    viewers_count = ?,
	    is_live = ?,
	    stream_key = ?,
	    playback_url = ?
	WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		stream.Title,
		stream.Description,
		stream.ThumbnailURL,
		stream.Category,
		stream.OwnerID,
		stream.ViewersCount,
		stream.IsLive,
		stream.StreamKey,
		stream.PlaybackURL,
		stream.ID,
	)

	return err
}