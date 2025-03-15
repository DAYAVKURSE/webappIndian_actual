package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/logger"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

// RouletteX14Sector defines the sector properties in the Roulette X14 game.
type RouletteX14Sector struct {
	Color        string `json:"color"`
	SectorId     int    `json:"sector_id"`
	SectorNumber int    `json:"sector_number"`
}

// RouletteX14BetInput defines the structure of a bet input.
type RouletteX14BetInput struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Color  string  `json:"color" binding:"required,oneof=black red green"`
}

// Predefined sectors on the Roulette X14 wheel.
var RouletteX14Sectors = []RouletteX14Sector{
	{Color: "red", SectorId: 1, SectorNumber: 1},
	{Color: "black", SectorId: 2, SectorNumber: 8},
	{Color: "red", SectorId: 3, SectorNumber: 2},
	{Color: "black", SectorId: 4, SectorNumber: 9},
	{Color: "red", SectorId: 5, SectorNumber: 3},
	{Color: "black", SectorId: 6, SectorNumber: 10},
	{Color: "red", SectorId: 7, SectorNumber: 4},
	{Color: "black", SectorId: 8, SectorNumber: 11},
	{Color: "red", SectorId: 9, SectorNumber: 5},
	{Color: "black", SectorId: 10, SectorNumber: 12},
	{Color: "red", SectorId: 11, SectorNumber: 6},
	{Color: "black", SectorId: 12, SectorNumber: 13},
	{Color: "red", SectorId: 13, SectorNumber: 7},
	{Color: "black", SectorId: 14, SectorNumber: 14},
	{Color: "green", SectorId: 15, SectorNumber: 0},
}

var (
	userLastBetTime      = make(map[int64]time.Time)
	userBetPattern       = make(map[int64][]string)
	userLastBetTimeMutex sync.Mutex
	userBetPatternMutex  sync.Mutex
	betCooldown          = 1 * time.Second
)

// PlaceRouletteX14Bet handles POST requests to place a bet on the Roulette X14 game.
func PlaceRouletteX14Bet(c *gin.Context) {
	var input RouletteX14BetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Check cooldown
	if !canPlaceBet(userID) {
		c.JSON(429, gin.H{"error": "Please wait before placing another bet"})
		return
	}

	result, err := processBet(userID, input)
	if err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			c.JSON(402, gin.H{"error": err.Error()})
		} else {
			logger.Error("%v", err)
			c.Status(500)
		}
		return
	}

	c.JSON(200, result)
}

func canPlaceBet(userID int64) bool {
	userLastBetTimeMutex.Lock()
	defer userLastBetTimeMutex.Unlock()

	lastBetTime, exists := userLastBetTime[userID]
	if !exists || time.Since(lastBetTime) >= betCooldown {
		userLastBetTime[userID] = time.Now()
		return true
	}
	return false
}

func processBet(userID int64, input RouletteX14BetInput) (gin.H, error) {
	var result gin.H

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		benefitFreeDeposit, applyBenefit, err := benefit_progress.UseFreeMiniGameBetIfAvailable(
			tx, userID, requirements.RouletteGameID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		var toCashBalance, toBonusBalance float64

		bet := models.RouletteX14Bet{
			UserID:       userID,
			IsBenefitBet: benefitFreeDeposit != 0,
			BetColor:     input.Color,
			CreatedAt:    time.Now(),
		}

		if benefitFreeDeposit == 0 {
			bonusBalance, err := exchange.GetUserExchangedBalanceAmount(tx, user.ID)
			if err != nil {
				return logger.WrapError(err, "")
			}

			if user.BalanceRupee+bonusBalance < input.Amount {
				return ErrInsufficientBalance
			}

			fromCashBalance, fromBonusBalance, err := exchange.UseExchangeBalancePayment(tx, &user, input.Amount)
			if err != nil {
				return logger.WrapError(err, "")
			}

			bet.Amount = fromCashBalance + fromBonusBalance
			bet.FromBonusBalance = fromBonusBalance
			bet.FromCashBalance = fromCashBalance
		} else {
			if err = applyBenefit(tx); err != nil {
				return logger.WrapError(err, "")
			}

			bet.Amount = benefitFreeDeposit
			bet.FromBonusBalance = benefitFreeDeposit
		}

		winningSector := spinRouletteX14Wheel(userID, input.Color)
		bet.Outcome = "lose"
		if bet.BetColor == winningSector.Color {
			if bet.BetColor == "green" {
				bet.Payout = bet.Amount * 14
				toCashBalance = bet.FromCashBalance * 14
				toBonusBalance = bet.FromBonusBalance * 14
			} else {
				bet.Payout = bet.Amount * 2
				toCashBalance = bet.FromCashBalance * 2
				toBonusBalance = bet.FromBonusBalance * 2
			}
			bet.Outcome = "win"
		}

		if err := tx.Create(&bet).Error; err != nil {
			return logger.WrapError(err, "")
		}

		// Store the game result
		gameResult := models.RouletteX14GameResult{
			UserID:       userID,
			WinningColor: winningSector.Color,
			SectorNumber: winningSector.SectorNumber,
			CreatedAt:    time.Now(),
		}
		if err := tx.Create(&gameResult).Error; err != nil {
			return logger.WrapError(err, "")
		}

		benefitWin := bet.Outcome == "win" && bet.IsBenefitBet
		if err = exchange.UpdateUserBalances(
			tx, &user, toCashBalance, toBonusBalance, benefitWin); err != nil {
			return logger.WrapError(err, "")
		}

		// win := models.Winning{
		// 	UserID:    user.ID,
		// 	WinAmount: toCashBalance + toBonusBalance,
		// }
	
		// if err := tx.Create(&win).Error; err != nil {
		// 	return logger.WrapError(err, "Failed to record winning")
		// }

		if !bet.IsBenefitBet {
			if err := updateRouletteGameTravePassLevelRequirement(
				tx, &bet, bet.FromCashBalance); err != nil {
				return logger.WrapError(err, "")
			}
		}

		result = gin.H{
			"bet_amount":     bet.Amount,
			"bet_color":      bet.BetColor,
			"outcome":        bet.Outcome,
			"payout":         bet.Payout,
			"winning_color":  winningSector.Color,
			"winning_number": winningSector.SectorNumber,
		}

		return nil
	})

	return result, err
}

func spinRouletteX14Wheel(userID int64, betColor string) RouletteX14Sector {
	userBetPatternMutex.Lock()
	defer userBetPatternMutex.Unlock()

	pattern, exists := userBetPattern[userID]
	if !exists {
		pattern = []string{}
	}

	pattern = append(pattern, betColor)
	if len(pattern) > 5 {
		pattern = pattern[1:]
	}
	userBetPattern[userID] = pattern

	if len(pattern) == 5 && pattern[0] == "black" && pattern[1] == "red" && pattern[2] == "red" && pattern[3] == "black" {
		// Reset pattern
		userBetPattern[userID] = []string{}
		// Return green sector
		return RouletteX14Sectors[14] // Green sector
	}

	return RouletteX14Sectors[rand.Intn(len(RouletteX14Sectors))]
}

func updateRouletteGameTravePassLevelRequirement(tx *gorm.DB, bet *models.RouletteX14Bet, fromCashBalance float64) error {
	requirementMGDone, err := requirement_progress.UpdateRequirementProgressMiniGameIfRequired(
		tx, bet.UserID, requirements.RouletteGameID, fromCashBalance, bet.Payout)
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementTurnoverDone, err := requirement_progress.UpdateRequirementProgressTurnoverIfRequired(
		tx, bet.UserID, fromCashBalance)
	if err != nil {
		return logger.WrapError(err, "")
	}

	if requirementMGDone || requirementTurnoverDone {
		if err = travepass.CheckAndUpgradeTravePassLevel(tx, bet.UserID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}

// GetRouletteX14Info returns the information for all sectors of the Roulette X14 wheel.
func GetRouletteX14Info(c *gin.Context) {
	c.JSON(200, RouletteX14Sectors)
}

func GetRouletteX14History(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var history []models.RouletteX14GameResult
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the latest 20 results
		if err := tx.Where("user_id = ?", userID).Order("created_at desc").Limit(20).Find(&history).Error; err != nil {
			return err
		}

		// If we have 20 results, delete older ones
		if len(history) == 20 {
			oldestTimestamp := history[len(history)-1].CreatedAt
			if err := tx.Where("user_id = ? AND created_at < ?", userID, oldestTimestamp).Delete(&models.RouletteX14GameResult{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.JSON(200, history)
}
