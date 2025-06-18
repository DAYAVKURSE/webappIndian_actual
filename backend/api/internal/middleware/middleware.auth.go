package middleware

import (
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const (
	TokenAccess  = "TokenAccess"
	TokenRefresh = "TokenRefresh"
	JWTkey       = "dasdasdasdasdas"
)

const (
	MiddlewareKeyAuth = "middleware_key_auth"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user id from context

		token, err := GetTokenFromAuthorizationHeader(c)
		if err != nil {
			logger.Error("%v", err)
			c.AbortWithStatus(400)
			return
		}

		userId, tokenType, err := TokenCheck(token, JWTkey)
		if err != nil {
			logger.Error("%v", err)
			if errors.Is(err, jwt.ErrTokenExpired) {
				logger.Error("%v", err)
				c.AbortWithStatus(401)
				return
			}
			logger.Error("%v", err)
			c.AbortWithStatus(400)
			return
		}

		if tokenType == TokenAccess {
			userIdSet(c, userId)
		} else {
			logger.Error("%v", err)
			c.AbortWithStatus(401)
		}

		// check if user in database
		exists, err := models.CheckIfUserExistsByID(int64(userId))
		if err != nil {
			logger.Error("%v", err)
			c.AbortWithStatus(500)
			return
		}

		// call c.Next if user in database
		// else response with 401
		if exists {
			c.Set(ContextUserIDKey, int64(userId))
			c.Next()
			return
		} else {
			c.JSON(401, gin.H{"error": "User not authorized"})
			c.Abort()
			return
		}

	}
}

func userIdSet(c *gin.Context, userId uint64) {
	//c.SetUserValue(MiddlewareKeyAuth, userId)

	//TODO подумать че то
}
