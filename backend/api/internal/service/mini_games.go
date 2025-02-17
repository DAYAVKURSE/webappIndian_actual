package service

import (
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserFreeMiniGameBets(c *gin.Context) {
	gameIDAny, ok := c.GetQuery("game_id")
	if !ok {
		c.JSON(400, gin.H{"error": "query parameter game_id invalid"})
		return
	}

	gameID, err := strconv.ParseInt(gameIDAny, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "query parameter game_id invalid"})
		return
	}

	if gameID != requirements.NvutiGameID &&
		gameID != requirements.DiceGameID &&
		gameID != requirements.RouletteGameID {
		c.JSON(400, gin.H{"error": "no game with this id"})
		return
	}

	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	benefitProgressMGs, err := benefit_progress.GetUserFreeMiniGameBets(nil, userID, gameID)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	if benefitProgressMGs == nil || len(*benefitProgressMGs) == 0 {
		c.String(404, "[]")
		return
	}

	c.JSON(200, *benefitProgressMGs)
}
