package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Carlvalencia1/streamhub-backend/internal/communities/domain"
	"github.com/google/uuid"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, c *domain.Community) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO communities (id, owner_id, name, description, image_url, invite_code, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		c.ID, c.OwnerID, c.Name, c.Description, c.ImageURL, c.InviteCode,
	)
	return err
}

func (r *MySQLRepository) GetByID(ctx context.Context, id string) (*domain.Community, error) {
	var c domain.Community
	err := r.db.QueryRowContext(ctx,
		`SELECT id, owner_id, name, description, image_url, invite_code, created_at, updated_at
		 FROM communities WHERE id = ?`, id,
	).Scan(&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.ImageURL, &c.InviteCode, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *MySQLRepository) GetByOwner(ctx context.Context, ownerID string) ([]*domain.Community, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, owner_id, name, description, image_url, invite_code, created_at, updated_at
		 FROM communities WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommunities(rows)
}

func (r *MySQLRepository) GetByMember(ctx context.Context, userID string) ([]*domain.Community, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.owner_id, c.name, c.description, c.image_url, c.invite_code, c.created_at, c.updated_at
		 FROM communities c
		 INNER JOIN community_members cm ON cm.community_id = c.id
		 WHERE cm.user_id = ? ORDER BY c.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommunities(rows)
}

func (r *MySQLRepository) GetByInviteCode(ctx context.Context, code string) (*domain.Community, error) {
	var c domain.Community
	err := r.db.QueryRowContext(ctx,
		`SELECT id, owner_id, name, description, image_url, invite_code, created_at, updated_at
		 FROM communities WHERE invite_code = ?`, code,
	).Scan(&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.ImageURL, &c.InviteCode, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *MySQLRepository) Update(ctx context.Context, c *domain.Community) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE communities SET name = ?, description = ?, image_url = ?, updated_at = NOW() WHERE id = ?`,
		c.Name, c.Description, c.ImageURL, c.ID,
	)
	return err
}

func (r *MySQLRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM communities WHERE id = ?`, id)
	return err
}

func (r *MySQLRepository) CreateChannel(ctx context.Context, ch *domain.Channel) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO community_channels (id, community_id, name, description, created_at)
		 VALUES (?, ?, ?, ?, NOW())`,
		ch.ID, ch.CommunityID, ch.Name, ch.Description,
	)
	return err
}

func (r *MySQLRepository) GetChannels(ctx context.Context, communityID string) ([]*domain.Channel, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, community_id, name, description, created_at
		 FROM community_channels WHERE community_id = ? ORDER BY created_at ASC`, communityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*domain.Channel
	for rows.Next() {
		var ch domain.Channel
		if err := rows.Scan(&ch.ID, &ch.CommunityID, &ch.Name, &ch.Description, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, &ch)
	}
	return channels, rows.Err()
}

func (r *MySQLRepository) UpdateChannel(ctx context.Context, ch *domain.Channel) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE community_channels SET name = ?, description = ? WHERE id = ?`,
		ch.Name, ch.Description, ch.ID,
	)
	return err
}

func (r *MySQLRepository) DeleteChannel(ctx context.Context, channelID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM community_channels WHERE id = ?`, channelID)
	return err
}

func (r *MySQLRepository) AddMember(ctx context.Context, communityID, userID, role string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO community_members (id, community_id, user_id, role, joined_at)
		 VALUES (?, ?, ?, ?, NOW())`,
		uuid.NewString(), communityID, userID, role,
	)
	return err
}

func (r *MySQLRepository) RemoveMember(ctx context.Context, communityID, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM community_members WHERE community_id = ? AND user_id = ?`,
		communityID, userID,
	)
	return err
}

func (r *MySQLRepository) GetMembers(ctx context.Context, communityID string) ([]*domain.Member, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT community_id, user_id, role, joined_at FROM community_members WHERE community_id = ?`, communityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.CommunityID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, rows.Err()
}

func (r *MySQLRepository) IsMember(ctx context.Context, communityID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community_members WHERE community_id = ? AND user_id = ?`,
		communityID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *MySQLRepository) SendMessage(ctx context.Context, msg *ChannelMessage) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO channel_messages (id, channel_id, user_id, type, content, media_url, poll_id, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		msg.ID, msg.ChannelID, msg.UserID, msg.Type, msg.Content, msg.MediaURL, msg.PollID, msg.ExpiresAt,
	)
	return err
}

func (r *MySQLRepository) GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*domain.ChannelMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT cm.id, cm.channel_id, cm.user_id, u.username, u.avatar_url,
		        cm.type, cm.content, cm.media_url, cm.poll_id, cm.expires_at, cm.created_at
		 FROM channel_messages cm
		 INNER JOIN users u ON u.id = cm.user_id
		 WHERE cm.channel_id = ? AND (cm.expires_at IS NULL OR cm.expires_at > NOW())
		 ORDER BY cm.created_at ASC
		 LIMIT ? OFFSET ?`,
		channelID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*domain.ChannelMessage
	for rows.Next() {
		var m domain.ChannelMessage
		if err := rows.Scan(
			&m.ID, &m.ChannelID, &m.UserID, &m.Username, &m.AvatarURL,
			&m.Type, &m.Content, &m.MediaURL, &m.PollID, &m.ExpiresAt, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, rows.Err()
}

func (r *MySQLRepository) DeleteMessage(ctx context.Context, messageID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channel_messages WHERE id = ?`, messageID)
	return err
}

func (r *MySQLRepository) DeleteExpiredMessages(ctx context.Context, channelID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM channel_messages WHERE channel_id = ? AND expires_at IS NOT NULL AND expires_at <= NOW()`,
		channelID,
	)
	return err
}

func (r *MySQLRepository) CreatePoll(ctx context.Context, poll *domain.Poll) error {
	optJSON, err := json.Marshal(poll.Options)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
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
	if err := json.Unmarshal([]byte(optJSON), &poll.Options); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT poll_id, user_id, option_index FROM poll_votes WHERE poll_id = ?`, pollID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var v domain.PollVote
		if err := rows.Scan(&v.PollID, &v.UserID, &v.OptionIndex); err != nil {
			return nil, err
		}
		poll.Votes = append(poll.Votes, v)
	}
	return &poll, rows.Err()
}

func (r *MySQLRepository) VotePoll(ctx context.Context, vote *domain.PollVote) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO poll_votes (id, poll_id, user_id, option_index) VALUES (?, ?, ?, ?)`,
		uuid.NewString(), vote.PollID, vote.UserID, vote.OptionIndex,
	)
	return err
}

func (r *MySQLRepository) GetChannelSettings(ctx context.Context, channelID string) (*domain.ChannelSettings, error) {
	var s domain.ChannelSettings
	err := r.db.QueryRowContext(ctx,
		`SELECT channel_id, disappearing_ttl_seconds FROM channel_settings WHERE channel_id = ?`, channelID,
	).Scan(&s.ChannelID, &s.DisappearingTTLSeconds)
	if err == sql.ErrNoRows {
		return &domain.ChannelSettings{ChannelID: channelID, DisappearingTTLSeconds: 0}, nil
	}
	return &s, err
}

func (r *MySQLRepository) UpsertChannelSettings(ctx context.Context, s *domain.ChannelSettings) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO channel_settings (channel_id, disappearing_ttl_seconds) VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE disappearing_ttl_seconds = VALUES(disappearing_ttl_seconds)`,
		s.ChannelID, s.DisappearingTTLSeconds,
	)
	return err
}

// SendMessage alias to satisfy interface (re-declares with domain type)
type ChannelMessage = domain.ChannelMessage

func (r *MySQLRepository) computeExpiry(channelID string, ctx context.Context) *time.Time {
	s, err := r.GetChannelSettings(ctx, channelID)
	if err != nil || s.DisappearingTTLSeconds == 0 {
		return nil
	}
	t := time.Now().Add(time.Duration(s.DisappearingTTLSeconds) * time.Second)
	return &t
}

func scanCommunities(rows *sql.Rows) ([]*domain.Community, error) {
	var list []*domain.Community
	for rows.Next() {
		var c domain.Community
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.ImageURL, &c.InviteCode, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}
