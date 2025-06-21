package service

import (
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const AccessExpiration = 10
const RefreshExpiration = 10

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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
	refreshExpiration := tmCreate + int64(RefreshExpiration*60*60)

	//TODO запись в бд токена

	refresh, err := middleware.TokenNew(middleware.JWTkey, user.ID, refreshExpiration, middleware.TokenRefresh)
	if err != nil {

		logger.Error(err.Error())
		c.AbortWithStatus(500)
		return
	}

	access, err := middleware.TokenNew(middleware.JWTkey, user.ID, accessExpiration, middleware.TokenAccess)
	if err != nil {

		logger.Error(err.Error())
		c.AbortWithStatus(500)
		return
	}

	token := Token{
		AccessToken:  access,
		RefreshToken: refresh,
	}

	c.JSON(200, token)
}

func RefreshLogin(c *gin.Context) {

	var req Token
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind  request: %v", err)
		c.JSON(400, gin.H{"error": "Invalid data"})
		return
	}
	logger.Info(req.RefreshToken)
	userId, tokenType, err := middleware.TokenCheck(req.RefreshToken, middleware.JWTkey)
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

	logger.Info(tokenType)
	if tokenType != middleware.TokenRefresh {
		logger.Error("%v", err)
		c.AbortWithStatus(400)
		return
	}

	//TODO GetRefreshTokenInfoByRefreshToken - получаем токен ищ база

	//TODO DeleteRefreshTokenById - удаляем токен
	tmCreate := time.Now().Unix()
	accessExpiration := tmCreate + int64(AccessExpiration*60*60)
	refreshExpiration := tmCreate + int64(RefreshExpiration*60*60)

	refresh, err := middleware.TokenNew(middleware.JWTkey, int64(userId), refreshExpiration, middleware.TokenRefresh)
	if err != nil {
		logger.Error(err.Error())
		c.AbortWithStatus(500)
		return
	}

	access, err := middleware.TokenNew(middleware.JWTkey, int64(userId), accessExpiration, middleware.TokenAccess)
	if err != nil {
		logger.Error(err.Error())
		c.AbortWithStatus(500)
		return
	}

	// TODO AddRefreshToken  добавляем токен
	token := Token{
		AccessToken:  access,
		RefreshToken: refresh,
	}

	c.JSON(200, token)
}
