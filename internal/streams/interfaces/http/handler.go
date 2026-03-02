package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/application"
)

type Handler struct {
	createUC *application.CreateStream
	getUC    *application.GetStreams
	startUC  *application.StartStream
	joinUC   *application.JoinStream
}

func NewHandler(
	createUC *application.CreateStream,
	getUC *application.GetStreams,
	startUC *application.StartStream,
	joinUC *application.JoinStream,
) *Handler {
	return &Handler{
		createUC: createUC,
		getUC:    getUC,
		startUC:  startUC,
		joinUC:   joinUC,
	}
}

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Category    string `json:"category"`
}

func (h *Handler) Create(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req createRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, stream)
}

func (h *Handler) GetAll(c *gin.Context) {

	streams, err := h.getUC.Execute(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get streams: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, streams)
}

func (h *Handler) Start(c *gin.Context) {

	id := c.Param("id")

	err := h.startUC.Execute(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start stream: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stream started"})
}

func (h *Handler) Join(c *gin.Context) {

	id := c.Param("id")

	err := h.joinUC.Execute(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join stream: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "joined stream successfully"})
}