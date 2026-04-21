package infrastructure

import (
	"context"
	"database/sql"

	"github.com/Carlvalencia1/streamhub-backend/internal/followers/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Follow(ctx context.Context, followerID, streamerID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO followers (follower_id, streamer_id) VALUES (?, ?)`,
		followerID, streamerID,
	)
	return err
}

func (r *MySQLRepository) Unfollow(ctx context.Context, followerID, streamerID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM followers WHERE follower_id = ? AND streamer_id = ?`,
		followerID, streamerID,
	)
	return err
}

func (r *MySQLRepository) IsFollowing(ctx context.Context, followerID, streamerID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM followers WHERE follower_id = ? AND streamer_id = ?`,
		followerID, streamerID,
	).Scan(&count)
	return count > 0, err
}

func (r *MySQLRepository) GetFollowerCount(ctx context.Context, streamerID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM followers WHERE streamer_id = ?`,
		streamerID,
	).Scan(&count)
	return count, err
}

func (r *MySQLRepository) GetFollowingIDs(ctx context.Context, followerID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT streamer_id FROM followers WHERE follower_id = ?`,
		followerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *MySQLRepository) GetFollowerUsers(ctx context.Context, streamerID string) ([]*domain.UserSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT u.id, u.username, u.nickname, u.avatar_url
		 FROM followers f
		 INNER JOIN users u ON u.id = f.follower_id
		 WHERE f.streamer_id = ?`,
		streamerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserSummaries(rows)
}

func (r *MySQLRepository) GetFollowingUsers(ctx context.Context, followerID string) ([]*domain.UserSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT u.id, u.username, u.nickname, u.avatar_url
		 FROM followers f
		 INNER JOIN users u ON u.id = f.streamer_id
		 WHERE f.follower_id = ?`,
		followerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserSummaries(rows)
}

func scanUserSummaries(rows *sql.Rows) ([]*domain.UserSummary, error) {
	var users []*domain.UserSummary
	for rows.Next() {
		var u domain.UserSummary
		if err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.AvatarURL); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *MySQLRepository) GetFollowerIDs(ctx context.Context, streamerID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT follower_id FROM followers WHERE streamer_id = ?`,
		streamerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
