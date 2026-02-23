package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	usersInfra "github.com/Carlvalencia1/streamhub-backend/internal/users/infrastructure"
	usersApp "github.com/Carlvalencia1/streamhub-backend/internal/users/application"
	usersHTTP "github.com/Carlvalencia1/streamhub-backend/internal/users/interfaces/http"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB) {

	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Users
	userRepo := usersInfra.NewMySQLRepository(db)

	registerUC := usersApp.NewRegisterUser(userRepo)
	loginUC := usersApp.NewLoginUser(userRepo)
	listUC := usersApp.NewListUsers(userRepo)

	handler := usersHTTP.NewHandler(registerUC, loginUC, listUC)

	usersHTTP.RegisterRoutes(api, handler)
}