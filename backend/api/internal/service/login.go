package service

import (
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

const AccessExpiration = 10

type Token struct {
	AccessToken string `json:"access_token"`
}

type Login struct {
	Nickname      string `json:"nickname"`
	Password      string `json:"password"`
	PasswordRetry string `json:"password_retry"`
}

func AuthLogin(c *gin.Context) {

	var req Login
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request: %v", err)
		c.JSON(400, gin.H{"error": "Invalid data"})
		return
	}

	user, err := models.GetUserWithPassword(req.Nickname)
	if err != nil {
		logger.Error("Failed get password: %v", err)
		c.JSON(400, gin.H{"error": "Invalid data"})
	}

	logger.Error(fmt.Sprintf("%v %v  user %v", req.Password, user.Password, user))

	if !middleware.ComparePasswords(user.Password, req.Password) {
		logger.Error("Error login or password incorrect")
		c.JSON(400, gin.H{"error": "Invalid data"})
		return
	}

	BaseAuth(c, &req, user)
}

func BaseAuth(c *gin.Context, req *Login, user *models.User) {
	req.Password = ""

	tmCreate := time.Now().Unix()
	accessExpiration := tmCreate + int64(AccessExpiration*60*60)

	access, err := middleware.TokenNew(middleware.JWTkey, user.ID, accessExpiration, middleware.TokenAccess)
	if err != nil {

		logger.Error(err.Error())
		c.AbortWithStatus(500)
		return
	}

	token := Token{
		AccessToken: access,
	}

	c.JSON(200, token)
}
