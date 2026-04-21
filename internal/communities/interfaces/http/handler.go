package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Carlvalencia1/streamhub-backend/internal/communities/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/communities/infrastructure"
)

type Handler struct {
	repo *infrastructure.MySQLRepository
}

func NewHandler(repo *infrastructure.MySQLRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) CreateCommunity(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
		ImageURL    *string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	community := &domain.Community{
		ID:          uuid.NewString(),
		OwnerID:     userID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		InviteCode:  uuid.NewString()[:8],
	}
	if err := h.repo.Create(c, community); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create community"})
		return
	}
	_ = h.repo.AddMember(c, community.ID, userID, "admin")
	c.JSON(http.StatusCreated, community)
}

func (h *Handler) GetMyCommunities(c *gin.Context) {
	userID := c.GetString("user_id")
	owned, _ := h.repo.GetByOwner(c, userID)
	member, _ := h.repo.GetByMember(c, userID)

	seen := map[string]bool{}
	var all []*domain.Community
	for _, co := range owned {
		seen[co.ID] = true
		all = append(all, co)
	}
	for _, co := range member {
		if !seen[co.ID] {
			all = append(all, co)
		}
	}
	if all == nil {
		all = []*domain.Community{}
	}
	c.JSON(http.StatusOK, gin.H{"communities": all})
}

func (h *Handler) GetCommunity(c *gin.Context) {
	id := c.Param("id")
	comm, err := h.repo.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
		return
	}
	channels, _ := h.repo.GetChannels(c, id)
	members, _ := h.repo.GetMembers(c, id)
	c.JSON(http.StatusOK, gin.H{
		"community": comm,
		"channels":  channels,
		"members":   members,
	})
}

func (h *Handler) UpdateCommunity(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	comm, err := h.repo.GetByID(c, id)
	if err != nil || comm.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		ImageURL    *string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != "" {
		comm.Name = req.Name
	}
	if req.Description != nil {
		comm.Description = req.Description
	}
	if req.ImageURL != nil {
		comm.ImageURL = req.ImageURL
	}
	if err := h.repo.Update(c, comm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	c.JSON(http.StatusOK, comm)
}

func (h *Handler) DeleteCommunity(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	comm, err := h.repo.GetByID(c, id)
	if err != nil || comm.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	_ = h.repo.Delete(c, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) JoinByInvite(c *gin.Context) {
	userID := c.GetString("user_id")
	code := c.Param("code")
	comm, err := h.repo.GetByInviteCode(c, code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	_ = h.repo.AddMember(c, comm.ID, userID, "member")
	c.JSON(http.StatusOK, gin.H{"community_id": comm.ID})
}

func (h *Handler) LeaveCommunity(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	_ = h.repo.RemoveMember(c, id, userID)
	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

func (h *Handler) CreateChannel(c *gin.Context) {
	userID := c.GetString("user_id")
	communityID := c.Param("id")
	comm, err := h.repo.GetByID(c, communityID)
	if err != nil || comm.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ch := &domain.Channel{
		ID:          uuid.NewString(),
		CommunityID: communityID,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := h.repo.CreateChannel(c, ch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
		return
	}
	c.JSON(http.StatusCreated, ch)
}

func (h *Handler) DeleteChannel(c *gin.Context) {
	userID := c.GetString("user_id")
	communityID := c.Param("id")
	channelID := c.Param("channelId")
	comm, err := h.repo.GetByID(c, communityID)
	if err != nil || comm.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	_ = h.repo.DeleteChannel(c, channelID)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) RemoveMember(c *gin.Context) {
	userID := c.GetString("user_id")
	communityID := c.Param("id")
	memberID := c.Param("memberId")
	comm, err := h.repo.GetByID(c, communityID)
	if err != nil || comm.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	_ = h.repo.RemoveMember(c, communityID, memberID)
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

func (h *Handler) GetMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	channelID := c.Param("channelId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	_ = h.repo.DeleteExpiredMessages(c, channelID)

	msgs, err := h.repo.GetMessages(c, channelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if msgs == nil {
		msgs = []*domain.ChannelMessage{}
	}

	myReactions, _ := h.repo.GetMyReactions(c, channelID, userID)
	for _, msg := range msgs {
		if msg.PollID != nil {
			if poll, err := h.repo.GetPoll(c, *msg.PollID); err == nil {
				msg.Poll = poll
			}
		}
		if emoji, ok := myReactions[msg.ID]; ok {
			msg.MyReaction = &emoji
		}
	}
	c.JSON(http.StatusOK, gin.H{"messages": msgs})
}

func (h *Handler) ReactToMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("messageId")
	var req struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.AddReaction(c, messageID, userID, req.Emoji); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"emoji": req.Emoji})
}

func (h *Handler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	channelID := c.Param("channelId")

	var req struct {
		Type    string  `json:"type"`
		Content string  `json:"content"`
		MediaURL *string `json:"media_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Type == "" {
		req.Type = "text"
	}

	settings, _ := h.repo.GetChannelSettings(c, channelID)
	var expiresAt *time.Time
	if settings != nil && settings.DisappearingTTLSeconds > 0 {
		t := time.Now().Add(time.Duration(settings.DisappearingTTLSeconds) * time.Second)
		expiresAt = &t
	}

	msg := &domain.ChannelMessage{
		ID:        uuid.NewString(),
		ChannelID: channelID,
		UserID:    userID,
		Type:      domain.MessageType(req.Type),
		Content:   req.Content,
		MediaURL:  req.MediaURL,
		ExpiresAt: expiresAt,
	}
	if err := h.repo.SendMessage(c, msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h *Handler) SendPollMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	channelID := c.Param("channelId")

	var req struct {
		Question       string   `json:"question" binding:"required"`
		Options        []string `json:"options" binding:"required"`
		MultipleChoice bool     `json:"multiple_choice"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least 2 options required"})
		return
	}

	poll := &domain.Poll{
		ID:             uuid.NewString(),
		Question:       req.Question,
		Options:        req.Options,
		MultipleChoice: req.MultipleChoice,
	}
	if err := h.repo.CreatePoll(c, poll); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	msg := &domain.ChannelMessage{
		ID:        uuid.NewString(),
		ChannelID: channelID,
		UserID:    userID,
		Type:      domain.MessageTypePoll,
		Content:   req.Question,
		PollID:    &poll.ID,
	}
	if err := h.repo.SendMessage(c, msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	msg.Poll = poll
	c.JSON(http.StatusCreated, msg)
}

func (h *Handler) VotePoll(c *gin.Context) {
	userID := c.GetString("user_id")
	pollID := c.Param("pollId")
	var req struct {
		OptionIndex int `json:"option_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	vote := &domain.PollVote{PollID: pollID, UserID: userID, OptionIndex: req.OptionIndex}
	if err := h.repo.VotePoll(c, vote); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	poll, _ := h.repo.GetPoll(c, pollID)
	c.JSON(http.StatusOK, poll)
}

func (h *Handler) DeleteMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	communityID := c.Param("id")
	messageID := c.Param("messageId")
	comm, err := h.repo.GetByID(c, communityID)
	if err != nil || comm.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	_ = h.repo.DeleteMessage(c, messageID)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) SetDisappearing(c *gin.Context) {
	userID := c.GetString("user_id")
	communityID := c.Param("id")
	channelID := c.Param("channelId")
	comm, err := h.repo.GetByID(c, communityID)
	if err != nil || comm.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req struct {
		TTLSeconds int `json:"ttl_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := &domain.ChannelSettings{ChannelID: channelID, DisappearingTTLSeconds: req.TTLSeconds}
	if err := h.repo.UpsertChannelSettings(c, s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}
