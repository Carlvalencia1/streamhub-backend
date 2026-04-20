package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/Carlvalencia1/streamhub-backend/internal/channelposts/domain"
	"github.com/google/uuid"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, post *domain.Post) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO channel_posts (id, streamer_id, type, content, media_url, poll_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, NOW())`,
		post.ID, post.StreamerID, post.Type, post.Content, post.MediaURL, post.PollID,
	)
	return err
}

func (r *MySQLRepository) GetByStreamer(ctx context.Context, streamerID string, limit, offset int) ([]*domain.Post, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT cp.id, cp.streamer_id, u.username, u.avatar_url, cp.type, cp.content, cp.media_url, cp.poll_id, cp.created_at
		 FROM channel_posts cp
		 INNER JOIN users u ON u.id = cp.streamer_id
		 WHERE cp.streamer_id = ?
		 ORDER BY cp.created_at DESC
		 LIMIT ? OFFSET ?`,
		streamerID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *MySQLRepository) GetFeed(ctx context.Context, followerID string, limit, offset int) ([]*domain.Post, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT cp.id, cp.streamer_id, u.username, u.avatar_url, cp.type, cp.content, cp.media_url, cp.poll_id, cp.created_at
		 FROM channel_posts cp
		 INNER JOIN users u ON u.id = cp.streamer_id
		 INNER JOIN followers f ON f.streamer_id = cp.streamer_id AND f.follower_id = ?
		 ORDER BY cp.created_at DESC
		 LIMIT ? OFFSET ?`,
		followerID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *MySQLRepository) Delete(ctx context.Context, postID, streamerID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM channel_posts WHERE id = ? AND streamer_id = ?`, postID, streamerID,
	)
	return err
}

func (r *MySQLRepository) CreatePoll(ctx context.Context, poll *domain.Poll) error {
	optJSON, _ := json.Marshal(poll.Options)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO polls (id, question, options, multiple_choice, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, NOW())`,
		poll.ID, poll.Question, string(optJSON), poll.MultipleChoice, poll.ExpiresAt,
	)
	return err
}

func (r *MySQLRepository) GetPoll(ctx context.Context, pollID string) (*domain.Poll, error) {
	var poll domain.Poll
	var optJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, question, options, multiple_choice, expires_at, created_at FROM polls WHERE id = ?`, pollID,
	).Scan(&poll.ID, &poll.Question, &optJSON, &poll.MultipleChoice, &poll.ExpiresAt, &poll.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(optJSON), &poll.Options)

	rows, _ := r.db.QueryContext(ctx,
		`SELECT poll_id, user_id, option_index FROM poll_votes WHERE poll_id = ?`, pollID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var v domain.PollVote
			rows.Scan(&v.PollID, &v.UserID, &v.OptionIndex)
			poll.Votes = append(poll.Votes, v)
		}
	}
	return &poll, nil
}

func (r *MySQLRepository) VotePoll(ctx context.Context, vote *domain.PollVote) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO poll_votes (id, poll_id, user_id, option_index) VALUES (?, ?, ?, ?)`,
		uuid.NewString(), vote.PollID, vote.UserID, vote.OptionIndex,
	)
	return err
}

func scanPosts(rows *sql.Rows) ([]*domain.Post, error) {
	var posts []*domain.Post
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.StreamerID, &p.Username, &p.AvatarURL,
			&p.Type, &p.Content, &p.MediaURL, &p.PollID, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}
	return posts, rows.Err()
}
