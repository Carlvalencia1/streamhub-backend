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

	followersInfra "github.com/Carlvalencia1/streamhub-backend/internal/followers/infrastructure"
	followersApp "github.com/Carlvalencia1/streamhub-backend/internal/followers/application"
	followersHTTP "github.com/Carlvalencia1/streamhub-backend/internal/followers/interfaces/http"

	communitiesInfra "github.com/Carlvalencia1/streamhub-backend/internal/communities/infrastructure"
	communitiesHTTP "github.com/Carlvalencia1/streamhub-backend/internal/communities/interfaces/http"

	channelpostsInfra "github.com/Carlvalencia1/streamhub-backend/internal/channelposts/infrastructure"
	channelpostsHTTP "github.com/Carlvalencia1/streamhub-backend/internal/channelposts/interfaces/http"

	uploadHTTP "github.com/Carlvalencia1/streamhub-backend/internal/upload/interfaces/http"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/config"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {

	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepo := usersInfra.NewMySQLRepository(db)

	registerUC := usersApp.NewRegisterUser(userRepo)
	loginUC := usersApp.NewLoginUser(userRepo)
	listUC := usersApp.NewListUsers(userRepo)
	googleAuthUC := usersApp.NewGoogleAuthUser(userRepo)

	handler := usersHTTP.NewHandler(registerUC, loginUC, listUC, googleAuthUC, userRepo)
	usersHTTP.RegisterRoutes(api, handler)

	// Protected routes
	protected := api.Group("/protected")
	protected.Use(middleware.AuthMiddleware())

	protected.GET("/me", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}
		uid := userID.(string)
		user, err := userRepo.GetByID(c, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
			return
		}

		var followersCount, followingCount int
		db.QueryRowContext(c, `SELECT COUNT(*) FROM followers WHERE streamer_id = ?`, uid).Scan(&followersCount)
		db.QueryRowContext(c, `SELECT COUNT(*) FROM followers WHERE follower_id = ?`, uid).Scan(&followingCount)

		c.JSON(200, gin.H{
			"user_id":         user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"role":            user.Role,
			"nickname":        user.Nickname,
			"bio":             user.Bio,
			"location":        user.Location,
			"avatar_url":      user.AvatarURL,
			"banner_url":      user.BannerURL,
			"followers_count": followersCount,
			"following_count": followingCount,
		})
	})

	protected.PUT("/role", handler.SetRole)

	type updateProfileRequest struct {
		Nickname  *string `json:"nickname"`
		Bio       *string `json:"bio"`
		Location  *string `json:"location"`
		BannerURL *string `json:"banner_url"`
	}
	protected.PUT("/profile", func(c *gin.Context) {
		userID := c.GetString("user_id")
		var req updateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := userRepo.UpdateProfile(c, userID, req.Nickname, req.Bio, req.Location); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
			return
		}
		if req.BannerURL != nil {
			_ = userRepo.UpdateBanner(c, userID, req.BannerURL)
		}
		c.JSON(200, gin.H{"message": "profile updated"})
	})

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

	validationHandler := streamsHTTP.NewStreamValidationHandler(streamsRepo)
	streamsHTTP.RegisterRoutes(api, streamsHandler, validationHandler)

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

	streamsApp.SetStreamLiveNotifier(notifyStreamLiveUC)

	followerRepo := followersInfra.NewMySQLRepository(db)

	followUC := followersApp.NewFollow(followerRepo)
	unfollowUC := followersApp.NewUnfollow(followerRepo)
	getStatusUC := followersApp.NewGetFollowerStatus(followerRepo)
	getFollowingUC := followersApp.NewGetFollowing(followerRepo)
	getFollowerUsersUC := followersApp.NewGetFollowerUsers(followerRepo)
	getFollowingUsersUC := followersApp.NewGetFollowingUsers(followerRepo)

	followersHandler := followersHTTP.NewHandler(followUC, unfollowUC, getStatusUC, getFollowingUC, getFollowerUsersUC, getFollowingUsersUC)
	followersHTTP.RegisterRoutes(api, followersHandler)

	communityRepo := communitiesInfra.NewMySQLRepository(db)
	communityHandler := communitiesHTTP.NewHandler(communityRepo)
	communitiesHTTP.RegisterRoutes(api, communityHandler)

	channelPostRepo := channelpostsInfra.NewMySQLRepository(db)
	channelPostHandler := channelpostsHTTP.NewHandler(channelPostRepo)
	channelpostsHTTP.RegisterRoutes(api, channelPostHandler)

	uploadHandler := uploadHTTP.NewHandler()
	uploadHTTP.RegisterRoutes(api, uploadHandler)

	r.Static("/uploads", "./uploads")
}
