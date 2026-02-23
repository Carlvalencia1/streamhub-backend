package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	usersInfra "github.com/Carlvalencia1/streamhub-backend/internal/users/infrastructure"
	usersApp "github.com/Carlvalencia1/streamhub-backend/internal/users/application"
	usersHTTP "github.com/Carlvalencia1/streamhub-backend/internal/users/interfaces/http"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB) {

	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})


	userRepo := usersInfra.NewMySQLRepository(db)

	registerUC := usersApp.NewRegisterUser(userRepo)
	loginUC := usersApp.NewLoginUser(userRepo)
	listUC := usersApp.NewListUsers(userRepo)

	handler := usersHTTP.NewHandler(registerUC, loginUC, listUC)

	usersHTTP.RegisterRoutes(api, handler)

	// =========================
	// Protected Routes
	// =========================

	protected := api.Group("/protected")
	protected.Use(middleware.AuthMiddleware())

	protected.GET("/me", func(c *gin.Context) {

		userID, _ := c.Get("userID")

		c.JSON(200, gin.H{
			"user_id": userID,
		})
	})
}