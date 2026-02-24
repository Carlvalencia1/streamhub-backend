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
}

func NewHandler(
	createUC *application.CreateStream,
	getUC *application.GetStreams,
	startUC *application.StartStream,
) *Handler {
	return &Handler{
		createUC: createUC,
		getUC:    getUC,
		startUC:  startUC,
	}
}

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Category    string `json:"category"`
}

func (h *Handler) Create(c *gin.Context) {

	userID, _ := c.Get("userID")

	var req createRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return
	}

	err := h.createUC.Execute(
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

	c.JSON(http.StatusCreated, gin.H{"message": "stream created"})
}

func (h *Handler) GetAll(c *gin.Context) {

	streams, err := h.getUC.Execute(c)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, streams)
}

func (h *Handler) Start(c *gin.Context) {

	id := c.Param("id")

	err := h.startUC.Execute(c, id)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stream started"})
}