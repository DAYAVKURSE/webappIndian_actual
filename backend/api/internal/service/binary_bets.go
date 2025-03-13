package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/binance"
	"BlessedApi/pkg/logger"
	"BlessedApi/pkg/redis"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type BinaryBetInput struct {
	Amount    float64                   `json:"amount" validate:"required,gt=0"`
	Duration  int64                     `json:"duration" validate:"required,min=10,max=600"`
	Direction models.BinaryBetDirection `json:"direction" validate:"required,oneof=up down"`
}

const WinMultiplier = 1.8 // 80% profit, shoudl adjust lates

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func PlaceBinaryBet(c *gin.Context, redisService *redis.RedisService) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var input BinaryBetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hasActiveBet, err := models.UserHasActiveBet(nil, userID)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}
	if hasActiveBet {
		c.JSON(400, gin.H{"error": "You have an active bet. Please wait for it to complete before placing a new one."})
		return
	}

	errInsufficientBalance := errors.New("insufficient balance")
	var bet models.BinaryBet

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Check user balance
		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		benefitFreeDeposit, applyBenefit, err :=
			benefit_progress.UseFreeBinaryOptionBetIfAvailable(tx, user.ID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		bonusBalance, err := exchange.GetUserExchangedBalanceAmount(tx, user.ID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		betAmount := benefitFreeDeposit
		if benefitFreeDeposit == 0 {
			if user.BalanceRupee+bonusBalance < input.Amount {
				return errInsufficientBalance
			}
			betAmount = input.Amount
		}

		currentPrice, err := getCurrentPrice(redisService)
		if err != nil {
			return logger.WrapError(err, "")
		}

		// Deduct bet amount from user balance
		var fromCashBalance, fromBonusBalance float64
		if benefitFreeDeposit == 0 {
			if fromCashBalance, fromBonusBalance, err = exchange.UseExchangeBalancePayment(
				tx, &user, input.Amount); err != nil {
				return logger.WrapError(err, "")
			}
		}

		now := time.Now()
		bet = models.BinaryBet{
			UserID:           userID,
			Amount:           betAmount,
			FromBonusBalance: fromBonusBalance,
			FromCashBalance:  fromCashBalance,
			IsBenefitBet:     benefitFreeDeposit != 0,
			Direction:        input.Direction,
			OpenedAt:         now,
			ExpiresAt:        now.Add(time.Duration(input.Duration) * time.Second),
			OpenPrice:        currentPrice,
		}

		if err := tx.Create(&bet).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if err = models.MaintainLastTenBinaryOptionBets(tx, userID); err != nil {
			return logger.WrapError(err, "")
		}

		if benefitFreeDeposit != 0 {
			if err = applyBenefit(tx); err != nil {
				return logger.WrapError(err, "")
			}
		}

		return nil
	})
	if err != nil && errors.Is(err, errInsufficientBalance) {
		c.JSON(402, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	go scheduleBetSettlement(
		bet.ID, time.Duration(input.Duration)*time.Second, redisService)

	c.JSON(200, bet)
}

func scheduleBetSettlement(betID int64, duration time.Duration, redisService *redis.RedisService) {
	time.Sleep(duration)
	err := settleBet(betID, redisService)
	if err != nil {
		logger.Error("Error settling bet %d: %v", betID, err)
		// In case of error, we just log it and ignore
		return
	}
}

func settleBet(betID int64, redisService *redis.RedisService) error {
	var bet models.BinaryBet
	if err := db.DB.First(&bet, betID).Error; err != nil {
		return logger.WrapError(err, "")
	}

	currentPrice, err := getCurrentPrice(redisService)
	if err != nil {
		return logger.WrapError(err, "")
	}

	bet.ClosePrice = currentPrice

	var payout float64
	if (bet.Direction == models.BetUp && bet.ClosePrice > bet.OpenPrice) ||
		(bet.Direction == models.BetDown && bet.ClosePrice < bet.OpenPrice) {
		bet.Outcome = "win"
		payout = bet.Amount * WinMultiplier
	} else {
		bet.Outcome = "loss"
		payout = 0
	}

	bet.Payout = payout

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&bet).Error; err != nil {
			return logger.WrapError(err, "")
		}

		var user models.User
		if err := tx.First(&user, "id = ?", bet.UserID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		var bonusWinAmount, cashWinAmount float64
		if bet.Outcome == "win" {
			bonusWinAmount = bet.FromBonusBalance * WinMultiplier
			cashWinAmount = bet.FromCashBalance * WinMultiplier
		}

		benefitWin := bet.Outcome == "win" && bet.IsBenefitBet
		err = exchange.UpdateUserBalances(tx, &user, cashWinAmount, bonusWinAmount, benefitWin)
		if err != nil {
			return logger.WrapError(err, "")
		}
		

		if !bet.IsBenefitBet {
			if err = updateBinaryOptionTravePassLevelRequirements(tx, &bet, cashWinAmount); err != nil {
				return logger.WrapError(err, "")
			}
		}

		return nil
	})
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

func updateBinaryOptionTravePassLevelRequirements(
	tx *gorm.DB, bet *models.BinaryBet, cashWinAmount float64) error {
	if tx == nil {
		tx = db.DB
	}

	var requirementBODone, requirementTurnoverDone bool
	requirementBODone, err := requirement_progress.UpdateRequirementProgressBinaryOptionIfRequired(
		tx, bet.UserID, bet.FromCashBalance, cashWinAmount)
	if err != nil {
		return logger.WrapError(err, "")
	}

	if bet.FromCashBalance > 0 {
		requirementTurnoverDone, err = requirement_progress.UpdateRequirementProgressTurnoverIfRequired(
			tx, bet.UserID, bet.FromCashBalance)
		if err != nil {
			return logger.WrapError(err, "")
		}

	}

	if requirementBODone || requirementTurnoverDone {
		if err = travepass.CheckAndUpgradeTravePassLevel(tx, bet.UserID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}

func GetUserBetOutcome(c *gin.Context) {
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

	var latestBets []models.BinaryBet
	if err := db.DB.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(10).
		Find(&latestBets).Error; err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	betResults := make([]gin.H, 0, len(latestBets))
	for _, bet := range latestBets {
		betResult := gin.H{
			"betID":      bet.ID,
			"amount":     bet.Amount,
			"direction":  bet.Direction,
			"openedAt":   bet.OpenedAt,
			"expiresAt":  bet.ExpiresAt,
			"openPrice":  bet.OpenPrice,
			"closePrice": bet.ClosePrice,
			"outcome":    bet.Outcome,
			"payout":     bet.Payout,
		}
		betResults = append(betResults, betResult)
	}

	c.JSON(200, gin.H{
		"userBalance": user.BalanceRupee,
		"latestBets":  betResults,
	})
}

func GetUserFreeBinaryOptionBets(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	benefitProgressBOs, err := benefit_progress.GetUserFreeBinaryOptionBets(nil, userID)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	if benefitProgressBOs == nil || len(*benefitProgressBOs) == 0 {
		c.String(404, "[]")
		return
	}

	c.JSON(200, *benefitProgressBOs)
}

func getCurrentPrice(redisService *redis.RedisService) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	latestKline, err := getLatestKline(ctx, redisService)
	if err != nil {
		return 0, logger.WrapError(err, "")
	}

	return latestKline.Close, nil
}

func getLatestKline(ctx context.Context, redisService *redis.RedisService) (binance.KlineData, error) {
	keys, err := fetchSortedKeys(ctx, redisService)
	if err != nil || len(keys) == 0 {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	klineData, err := fetchSingleKlineData(ctx, keys[len(keys)-1], redisService)
	if err != nil {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	return klineData, nil
}

func fetchSortedKeys(ctx context.Context, redisService *redis.RedisService) ([]string, error) {
	keys, err := redisService.Client().Keys(ctx, "binance_kline_*").Result()
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	if len(keys) == 0 {
		return nil, logger.WrapError(fmt.Errorf("no kline data found in Redis"), "")
	}

	sort.Strings(keys)
	return keys, nil
}

func fetchSingleKlineData(ctx context.Context, key string, redisService *redis.RedisService) (binance.KlineData, error) {
	data, err := redisService.GetKey(ctx, key)
	if err != nil {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	var kline binance.KlineData
	if err := json.Unmarshal([]byte(data), &kline); err != nil {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	return kline, nil
}
