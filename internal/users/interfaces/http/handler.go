package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/users/application"
)

type Handler struct {
	registerUC *application.RegisterUser
	loginUC    *application.LoginUser
	listUC     *application.ListUsers
}

func NewHandler(
	registerUC *application.RegisterUser,
	loginUC *application.LoginUser,
	listUC *application.ListUsers,
) *Handler {
	return &Handler{
		registerUC: registerUC,
		loginUC:    loginUC,
		listUC:     listUC,
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *gin.Context) {

	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.registerUC.Execute(c, req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (h *Handler) Login(c *gin.Context) {

	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token, err := h.loginUC.Execute(c, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) List(c *gin.Context) {

	users, err := h.listUC.Execute(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Map to UserResponse without password
	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}