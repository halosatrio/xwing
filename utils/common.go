package utils

import "github.com/gin-gonic/gin"

func RespondError(c *gin.Context, status int, message, err string) any {
	c.JSON(status, gin.H{
		"status":  status,
		"message": message,
		"error":   err,
	})
	return nil
}
