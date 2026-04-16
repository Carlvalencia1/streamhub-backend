package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {

	users := r.Group("/users")

	users.POST("/register", handler.Register)
	users.POST("/login", handler.Login)
	users.POST("/google-auth", handler.GoogleAuth)
	users.GET("", handler.List)
}