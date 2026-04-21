package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Carlvalencia1/streamhub-backend/internal/channelposts/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/channelposts/infrastructure"
)

type Handler struct {
	repo *infrastructure.MySQLRepository
}

func NewHandler(repo *infrastructure.MySQLRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) CreatePost(c *gin.Context) {
	streamerID := c.GetString("user_id")
	var req struct {
		Type     string  `json:"type"`
		Content  string  `json:"content"`
		MediaURL *string `json:"media_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Type == "" {
		req.Type = "text"
	}
	post := &domain.Post{
		ID:         uuid.NewString(),
		StreamerID: streamerID,
		Type:       domain.PostType(req.Type),
		Content:    req.Content,
		MediaURL:   req.MediaURL,
	}
	if err := h.repo.Create(c, post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, post)
}

func (h *Handler) CreatePollPost(c *gin.Context) {
	streamerID := c.GetString("user_id")
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
	post := &domain.Post{
		ID:         uuid.NewString(),
		StreamerID: streamerID,
		Type:       domain.PostTypePoll,
		Content:    req.Question,
		PollID:     &poll.ID,
	}
	if err := h.repo.Create(c, post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	post.Poll = poll
	c.JSON(http.StatusCreated, post)
}

func (h *Handler) GetMyPosts(c *gin.Context) {
	streamerID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	posts, err := h.repo.GetByStreamer(c, streamerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if posts == nil {
		posts = []*domain.Post{}
	}
	h.hydratePosts(c, posts)
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func (h *Handler) GetStreamerPosts(c *gin.Context) {
	streamerID := c.Param("streamerId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	posts, err := h.repo.GetByStreamer(c, streamerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if posts == nil {
		posts = []*domain.Post{}
	}
	h.hydratePosts(c, posts)
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func (h *Handler) GetFeed(c *gin.Context) {
	followerID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	posts, err := h.repo.GetFeed(c, followerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if posts == nil {
		posts = []*domain.Post{}
	}
	h.hydratePosts(c, posts)
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func (h *Handler) DeletePost(c *gin.Context) {
	streamerID := c.GetString("user_id")
	postID := c.Param("postId")
	_ = h.repo.Delete(c, postID, streamerID)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
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

func (h *Handler) hydratePosts(c *gin.Context, posts []*domain.Post) {
	for _, p := range posts {
		if p.PollID != nil {
			poll, err := h.repo.GetPoll(c, *p.PollID)
			if err == nil {
				p.Poll = poll
			}
		}
	}
}
