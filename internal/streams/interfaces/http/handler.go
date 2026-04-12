package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/Carlvalencia1/streamhub-backend/internal/streams/application"
	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

type Handler struct {
	createUC    *application.CreateStream
	getUC       *application.GetStreams
	getByIDUC   *application.GetStreamByID
	startUC     *application.StartStream
	stopUC      *application.StopStream
	joinUC      *application.JoinStream
}

func NewHandler(
	createUC *application.CreateStream,
	getUC *application.GetStreams,
	getByIDUC *application.GetStreamByID,
	startUC *application.StartStream,
	stopUC *application.StopStream,
	joinUC *application.JoinStream,
) *Handler {
	return &Handler{
		createUC:  createUC,
		getUC:     getUC,
		getByIDUC: getByIDUC,
		startUC:   startUC,
		stopUC:    stopUC,
		joinUC:    joinUC,
	}
}

type createRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Thumbnail   string `json:"thumbnail" binding:"required,url"`
	Category    string `json:"category" binding:"required,min=3,max=100"`
}

type createResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	StreamKey   string `json:"stream_key"`
	RTMPUrl     string `json:"rtmp_url"`
	PlaybackURL string `json:"playback_url"`
}

func (h *Handler) Create(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("user_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req createRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.ErrorWithContext("CreateStream", "invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validaciones de negocio
	if strings.TrimSpace(req.Title) == "" {
		logger.Warn("empty title provided by user " + userID.(string))
		c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot be empty"})
		return
	}

	if strings.TrimSpace(req.Category) == "" {
		logger.Warn("empty category provided by user " + userID.(string))
		c.JSON(http.StatusBadRequest, gin.H{"error": "category cannot be empty"})
		return
	}

	stream, err := h.createUC.Execute(
		c,
		req.Title,
		req.Description,
		req.Thumbnail,
		req.Category,
		userID.(string),
	)

	if err != nil {
		logger.ErrorWithContext("CreateStream", "failed to create stream", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create stream"})
		return
	}

	logger.StreamEvent("CREATED", stream.ID, fmt.Sprintf("Title: %s | Owner: %s", stream.Title, userID.(string)))

	rtmpURL := "rtmp://54.144.66.251/live/" + stream.StreamKey

	response := createResponse{
		ID:          stream.ID,
		Title:       stream.Title,
		StreamKey:   stream.StreamKey,
		RTMPUrl:     rtmpURL,
		PlaybackURL: stream.PlaybackURL,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetAll(c *gin.Context) {

	streams, err := h.getUC.Execute(c)
	if err != nil {
		logger.ErrorWithContext("GetAll", "failed to retrieve all streams", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve streams"})
		return
	}

	if streams == nil {
		streams = []*domain.Stream{}
	}

	logger.Debug(fmt.Sprintf("retrieved %d streams", len(streams)))
	c.JSON(http.StatusOK, streams)
}

func (h *Handler) GetByID(c *gin.Context) {

	id := c.Param("id")

	if strings.TrimSpace(id) == "" {
		logger.Warn("empty stream id for get")
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream id is required"})
		return
	}

	stream, err := h.getByIDUC.Execute(c, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("stream not found: " + id)
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		logger.ErrorWithContext("GetByID", "failed to get stream "+id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve stream"})
		return
	}

	c.JSON(http.StatusOK, stream)
}

func (h *Handler) Start(c *gin.Context) {

	id := c.Param("id")

	if strings.TrimSpace(id) == "" {
		logger.Warn("empty stream id for start")
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream id is required"})
		return
	}

	err := h.startUC.Execute(c, id)
	if err != nil {
		logger.ErrorWithContext("StartStream", "failed to start stream "+id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start stream"})
		return
	}

	logger.StreamEvent("STARTED", id, "Stream went live")
	c.JSON(http.StatusOK, gin.H{"message": "stream started successfully"})
}

func (h *Handler) Stop(c *gin.Context) {

	id := c.Param("id")

	if strings.TrimSpace(id) == "" {
		logger.Warn("empty stream id for stop")
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream id is required"})
		return
	}

	err := h.stopUC.Execute(c, id)
	if err != nil {
		logger.ErrorWithContext("StopStream", "failed to stop stream "+id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop stream"})
		return
	}

	logger.StreamEvent("STOPPED", id, "Stream ended")
	c.JSON(http.StatusOK, gin.H{"message": "stream stopped successfully"})
}

func (h *Handler) Join(c *gin.Context) {

	id := c.Param("id")

	if strings.TrimSpace(id) == "" {
		logger.Warn("empty stream id for join")
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream id is required"})
		return
	}

	err := h.joinUC.Execute(c, id)
	if err != nil {
		logger.ErrorWithContext("JoinStream", "failed to join stream "+id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join stream"})
		return
	}

	logger.StreamEvent("USER_JOINED", id, "New viewer joined")
	c.JSON(http.StatusOK, gin.H{"message": "joined stream successfully"})
}