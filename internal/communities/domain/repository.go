package domain

import "context"

type Repository interface {
	Create(ctx context.Context, c *Community) error
	GetByID(ctx context.Context, id string) (*Community, error)
	GetByOwner(ctx context.Context, ownerID string) ([]*Community, error)
	GetByMember(ctx context.Context, userID string) ([]*Community, error)
	GetByInviteCode(ctx context.Context, code string) (*Community, error)
	Update(ctx context.Context, c *Community) error
	Delete(ctx context.Context, id string) error

	CreateChannel(ctx context.Context, ch *Channel) error
	GetChannels(ctx context.Context, communityID string) ([]*Channel, error)
	UpdateChannel(ctx context.Context, ch *Channel) error
	DeleteChannel(ctx context.Context, channelID string) error

	AddMember(ctx context.Context, communityID, userID, role string) error
	RemoveMember(ctx context.Context, communityID, userID string) error
	GetMembers(ctx context.Context, communityID string) ([]*Member, error)
	IsMember(ctx context.Context, communityID, userID string) (bool, error)

	SendMessage(ctx context.Context, msg *ChannelMessage) error
	GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*ChannelMessage, error)
	DeleteMessage(ctx context.Context, messageID string) error
	DeleteExpiredMessages(ctx context.Context, channelID string) error

	CreatePoll(ctx context.Context, poll *Poll) error
	GetPoll(ctx context.Context, pollID string) (*Poll, error)
	VotePoll(ctx context.Context, vote *PollVote) error

	GetChannelSettings(ctx context.Context, channelID string) (*ChannelSettings, error)
	UpsertChannelSettings(ctx context.Context, s *ChannelSettings) error
}
