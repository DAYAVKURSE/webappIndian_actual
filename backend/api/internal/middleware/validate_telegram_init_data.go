package middleware

import (
	"BlessedApi/pkg/logger"
	"errors"
	"log" // Добавляем стандартный логгер
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

const (
	ContextUserIDKey   = "user_id"
	InitDataExpiration = 24 * time.Hour
)

var telegramBotToken string

func init() {
	var ok bool
	telegramBotToken, ok = os.LookupEnv("TOKEN")
	if !ok {
		logger.Fatal("unable to get telegram bot token from environment")
	}
}

func ValidateTelegramInitDataMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var initData string

		// Check if it's a WebSocket upgrade request
		if c.IsWebsocket() {
			// For WebSocket connections, get init data from query parameter
			initData = c.Query("init_data")
		} else {
			// For regular HTTP requests, get init data from header
			initData = c.GetHeader("X-Telegram-Init-Dat1a")
		}

		if initData == "" {
			c.JSON(400, gin.H{"error": "Missing Telegram init data"})
			c.Abort()
			return
		}

		// Rest of the validation logic
		err := initdata.Validate(initData, telegramBotToken, InitDataExpiration)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		parsedData, err := initdata.Parse(initData)
		if err != nil {
			c.JSON(400, gin.H{"error": "Failed to parse Telegram init data"})
			c.Abort()
			return
		}

		if parsedData.User.ID == 0 {
			c.JSON(400, gin.H{"error": "User ID is zero"})
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, parsedData.User.ID)
		c.Next()
	}
}

func GetUserIDFromGinContext(c *gin.Context) (int64, error) {
	// Get user_id from middleware
	userIDAny, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, logger.WrapError(errors.New("user_id not in GIN context"), "")
	}

	userIDInt, ok := userIDAny.(int64)
	if !ok {
		return 0, logger.WrapError(errors.New("unable to cast user_id value to int"), "")
	}

	log.Printf("GetUserIDFromGinContext - checking context keys: %+v", c.Keys)
	logger.Warn(strconv.FormatInt(userIDInt, 10))

	return userIDInt, nil
}
