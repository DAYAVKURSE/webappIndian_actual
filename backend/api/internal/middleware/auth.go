package middleware

import (
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user id from context
		userID, err := GetUserIDFromGinContext(c)
		if err != nil {
			logger.Error("%v", err)
			c.AbortWithStatus(500)
			return
		}

		// check if user in database
		exists, err := models.CheckIfUserExistsByID(userID)
		if err != nil {
			logger.Error("%v", err)
			c.AbortWithStatus(500)
			return
		}

		// call c.Next if user in database
		// else response with 401
		if exists {
			c.Next()
			return
		} else {
			c.JSON(401, gin.H{"error": "User not authorized"})
			c.Abort()
			return
		}
	}
}
