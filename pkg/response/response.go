package response

import "github.com/gin-gonic/gin"

// SuccessResponse es la estructura estándar para respuestas exitosas
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse es la estructura estándar para respuestas de error
type ErrorResponse struct {
	Error string `json:"error"`
}

func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"data": data,
	})
}

func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{
		"error": msg,
	})
}