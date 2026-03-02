package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	usersInfra "github.com/Carlvalencia1/streamhub-backend/internal/users/infrastructure"
	usersApp "github.com/Carlvalencia1/streamhub-backend/internal/users/application"
	usersHTTP "github.com/Carlvalencia1/streamhub-backend/internal/users/interfaces/http"

	streamsInfra "github.com/Carlvalencia1/streamhub-backend/internal/streams/infrastructure"
	streamsApp "github.com/Carlvalencia1/streamhub-backend/internal/streams/application"
	streamsHTTP "github.com/Carlvalencia1/streamhub-backend/internal/streams/interfaces/http"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB) {

	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// =========================
	// Users Module
	// =========================

	userRepo := usersInfra.NewMySQLRepository(db)

	registerUC := usersApp.NewRegisterUser(userRepo)
	loginUC := usersApp.NewLoginUser(userRepo)
	listUC := usersApp.NewListUsers(userRepo)

	handler := usersHTTP.NewHandler(registerUC, loginUC, listUC)

	usersHTTP.RegisterRoutes(api, handler)

	// =========================
	// Streams Module
	// =========================

	streamsRepo := streamsInfra.NewMySQLRepository(db)

	createStreamUC := streamsApp.NewCreateStream(streamsRepo)
	getStreamsUC := streamsApp.NewGetStreams(streamsRepo)
	startStreamUC := streamsApp.NewStartStream(streamsRepo)
	joinStreamUC := streamsApp.NewJoinStream(streamsRepo)

	streamsHandler := streamsHTTP.NewHandler(
		createStreamUC,
		getStreamsUC,
		startStreamUC,
		joinStreamUC,
	)

	streamsHTTP.RegisterRoutes(api, streamsHandler)

	// =========================
	// Protected Routes Example
	// =========================

	protected := api.Group("/protected")
	protected.Use(middleware.AuthMiddleware())

	protected.GET("/me", func(c *gin.Context) {

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		c.JSON(200, gin.H{
			"user_id": userID,
		})
	})
}