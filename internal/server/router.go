package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	usersInfra "github.com/Carlvalencia1/streamhub-backend/internal/users/infrastructure"
	usersApp "github.com/Carlvalencia1/streamhub-backend/internal/users/application"
	usersHTTP "github.com/Carlvalencia1/streamhub-backend/internal/users/interfaces/http"

	streamsInfra "github.com/Carlvalencia1/streamhub-backend/internal/streams/infrastructure"
	streamsApp "github.com/Carlvalencia1/streamhub-backend/internal/streams/application"
	streamsHTTP "github.com/Carlvalencia1/streamhub-backend/internal/streams/interfaces/http"

	notificationsInfra "github.com/Carlvalencia1/streamhub-backend/internal/notifications/infrastructure"
	notificationsApp "github.com/Carlvalencia1/streamhub-backend/internal/notifications/application"
	notificationsHTTP "github.com/Carlvalencia1/streamhub-backend/internal/notifications/interfaces/http"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/config"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {

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
	googleAuthUC := usersApp.NewGoogleAuthUser(userRepo)

	handler := usersHTTP.NewHandler(registerUC, loginUC, listUC, googleAuthUC)

	usersHTTP.RegisterRoutes(api, handler)

	// =========================
	// Streams Module
	// =========================

	streamsRepo := streamsInfra.NewMySQLRepository(db)

	createStreamUC := streamsApp.NewCreateStream(streamsRepo)
	getStreamsUC := streamsApp.NewGetStreams(streamsRepo)
	getStreamByIDUC := streamsApp.NewGetStreamByID(streamsRepo)
	startStreamUC := streamsApp.NewStartStream(streamsRepo)
	stopStreamUC := streamsApp.NewStopStream(streamsRepo)
	joinStreamUC := streamsApp.NewJoinStream(streamsRepo)

	streamsHandler := streamsHTTP.NewHandler(
		createStreamUC,
		getStreamsUC,
		getStreamByIDUC,
		startStreamUC,
		stopStreamUC,
		joinStreamUC,
	)

	// Validation handler (for NGINX RTMP webhooks)
	validationHandler := streamsHTTP.NewStreamValidationHandler(streamsRepo)

	streamsHTTP.RegisterRoutes(api, streamsHandler, validationHandler)

	// =========================
	// Notifications Module (FCM)
	// =========================

	// Inicializar Firebase Push Provider
	var firebasePushProvider *notificationsInfra.FirebasePushProvider
	if cfg.FirebaseCredentialsPath != "" {
		var err error
		firebasePushProvider, err = notificationsInfra.NewFirebasePushProvider(cfg.FirebaseCredentialsPath)
		if err != nil {
			logger.Error("failed to initialize Firebase provider: " + err.Error())
			log.Fatalf("Failed to initialize Firebase: %v", err)
		}
	} else {
		logger.Warn("FIREBASE_CREDENTIALS_PATH not set, push notifications will be disabled")
	}

	notificationRepo := notificationsInfra.NewDeviceTokenRepository(db)

	registerTokenUC := notificationsApp.NewRegisterFcmToken(notificationRepo)
	removeTokenUC := notificationsApp.NewRemoveFcmToken(notificationRepo)
	notifyStreamLiveUC := notificationsApp.NewNotifyStreamLive(notificationRepo, firebasePushProvider)

	notificationHandler := notificationsHTTP.NewHandler(registerTokenUC, removeTokenUC)

	notificationsHTTP.RegisterRoutes(api, notificationHandler)

	// Inyectar notifyStreamLiveUC en streams module (para usarlo en StartStream)
	streamsApp.SetStreamLiveNotifier(notifyStreamLiveUC)

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