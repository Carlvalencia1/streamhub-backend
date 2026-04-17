package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/users/application"
)

type Handler struct {
	registerUC   *application.RegisterUser
	loginUC      *application.LoginUser
	listUC       *application.ListUsers
	googleAuthUC *application.GoogleAuthUser
	userRepo     application.UserRepo
}

func NewHandler(
	registerUC *application.RegisterUser,
	loginUC *application.LoginUser,
	listUC *application.ListUsers,
	googleAuthUC *application.GoogleAuthUser,
	userRepo application.UserRepo,
) *Handler {
	return &Handler{
		registerUC:   registerUC,
		loginUC:      loginUC,
		listUC:       listUC,
		googleAuthUC: googleAuthUC,
		userRepo:     userRepo,
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "viewer"
	}
	err := h.registerUC.Execute(c, req.Username, req.Email, req.Password, req.Role)
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

type googleAuthRequest struct {
	IDToken string `json:"id_token"`
}

func (h *Handler) GoogleAuth(c *gin.Context) {
	var req googleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.IDToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token required"})
		return
	}
	token, err := h.googleAuthUC.Execute(c, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) List(c *gin.Context) {
	users, err := h.listUC.Execute(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, response)
}

type setRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=streamer viewer"`
}

func (h *Handler) SetRole(c *gin.Context) {
	userID := c.GetString("user_id")
	var req setRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be 'streamer' or 'viewer'"})
		return
	}
	if err := h.userRepo.UpdateRole(c, userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": req.Role})
}
