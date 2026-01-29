package utils

import (
	"github.com/gin-gonic/gin"
)

func RespondError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
