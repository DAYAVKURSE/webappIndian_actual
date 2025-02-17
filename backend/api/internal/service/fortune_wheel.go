package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/internal/models/fortune_wheel"
	"BlessedApi/pkg/logger"
	"BlessedApi/pkg/redis"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var redisService *redis.RedisService

func InitFortuneWheelService(rs *redis.RedisService) {
	redisService = rs
}

// WinData represents a single Fortune Wheel win
type FortuneWheelWinData struct {
	Nickname           string
	FortuneWheelSector fortune_wheel.FortuneWheelSector
	Timestamp          int64
}

// GetFortuneWheelInfo returns the information for all sectors of the fortune wheel
func GetFortuneWheelInfo(c *gin.Context) {
	fWSectors, err := fortune_wheel.GetAllFortuneWheelSectors(nil)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.JSON(200, fWSectors)
}

func GetUserFortuneWheelAvailableSpins(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var totalSpins int
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		totalSpins, err = benefit_progress.GetUserAvailableSpins(nil, userID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{"TotalSpins": totalSpins})
}

func SpinFortuneWheel(c *gin.Context) {
	if redisService == nil {
		logger.Error("Redis service is not initialized")
		c.Status(500)
		return
	}

	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Update spins and balance in a transaction
	errNoSpinsAvailable := errors.New("no spins available")
	var selectedSector fortune_wheel.FortuneWheelSector

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		isSpinsAvailable, applyBenefit, err :=
			benefit_progress.UseFreeFortuneWheelSpinIfAvailable(tx, userID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		if !isSpinsAvailable {
			return errNoSpinsAvailable
		}

		selectedSector, err = fortune_wheel.WeightedRandomSelection(tx)
		if err != nil {
			return logger.WrapError(err, "")
		}

		if err = applyBenefit(tx); err != nil {
			return logger.WrapError(err, "")
		}

		// Apply sector benefits
		if err := benefit_progress.CreateOrApplyPolymorphicBenefitProgress(
			tx, &selectedSector.Benefit, userID); err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil && errors.Is(err, errNoSpinsAvailable) {
		c.JSON(403, gin.H{"error": "No spins available"})
		return
	} else if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Store win data in Redis for websocket service

	winData := FortuneWheelWinData{
		Nickname:           user.Nickname,
		FortuneWheelSector: selectedSector,
		Timestamp:          time.Now().UnixNano() / int64(time.Millisecond),
	}

	winDataJSON, err := json.Marshal(winData)
	if err != nil {
		logger.Error("Failed to marshal win data: %v", err)
		c.Status(500)
		return
	}

	// Generate a unique key for win
	winKey := fmt.Sprintf("fortune_wheel:win:%d", winData.Timestamp)
	ctx := context.Background()
	err = redisService.SetKey(ctx, winKey, string(winDataJSON), 1*time.Minute)
	if err != nil {
		logger.Error("%v", err)
	}

	c.JSON(200, selectedSector)
}

// This is ONLY for testing purposes
type AddSpinsInput struct {
	UserID int64 `json:"UserID" binding:"required"`
	Spins  int   `json:"Spins" binding:"required"`
}

func AddSpins(c *gin.Context) {
	var input AddSpinsInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.Status(403)
		return
	}

	var benefit benefits.Benefit
	var totalSpins int

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		benefitFortuneWheel := benefits.BenefitFortuneWheel{
			FreeSpinsAmount: input.Spins}
		if err = tx.Create(&benefitFortuneWheel).
			Scan(&benefitFortuneWheel).Error; err != nil {
			return logger.WrapError(err, "")
		}

		benefit = benefits.Benefit{
			PolymorphicBenefitID:   benefitFortuneWheel.ID,
			PolymorphicBenefitType: benefits.BenefitFortuneWheelType}
		if err = tx.Create(&benefit).Error; err != nil {
			return logger.WrapError(err, "")
		}
		benefit.PolymorphicBenefit = benefitFortuneWheel

		if err = benefit_progress.CreateOrApplyPolymorphicBenefitProgress(
			tx, &benefit, input.UserID); err != nil {
			return logger.WrapError(err, "")
		}

		totalSpins, err = benefit_progress.GetUserAvailableSpins(tx, input.UserID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil {
		logger.WrapError(err, "")
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{"Message": "Spins added successfully", "TotalSpins": totalSpins})
}
