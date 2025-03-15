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

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DiceBetDirection string

const (
	DiceBetLess      DiceBetDirection = "less"
	DiceBetNvutiMore DiceBetDirection = "more"
)

type DiceBetInput struct {
	Amount     float64          `json:"amount" validate:"required,gt=0"`
	WinPercent int64            `json:"winPercent" validate:"required,min=1,max=95"`
	Direction  DiceBetDirection `json:"direction" validate:"required,oneof=more less"`
}

type DiceBetResult struct {
	Won    bool `json:"won"`
	Number int  `json:"number"`
}

func DicePlaceBet(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var input DiceBetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Status(400)
		return
	}

	if err := validate.Struct(input); err != nil {
		c.Status(400)
		return
	}

	errInsufficientBalance := errors.New("insufficient balance")
	var result DiceBetResult
	var toCashBalance, toBonusBalance float64
	var fromCashBalance, fromBonusBalance float64

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		// Get user free bets
		benefitFreeDeposit, applyBenefit, err :=
			benefit_progress.UseFreeMiniGameBetIfAvailable(tx, userID, requirements.DiceGameID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		// If there are no free bets
		if benefitFreeDeposit == 0 {
			bonusBalance, err := exchange.GetUserExchangedBalanceAmount(tx, user.ID)
			if err != nil {
				return logger.WrapError(err, "")
			}

			// User dont have enough money on both balances
			if user.BalanceRupee+bonusBalance < input.Amount {
				return errInsufficientBalance
			}

			// Pay with mixed balances
			fromCashBalance, fromBonusBalance, err = exchange.UseExchangeBalancePayment(tx, &user, input.Amount)
			if err != nil {
				return logger.WrapError(err, "")
			}

			// Calculate winnings and settle result
			multiplier := 100 / float64(input.WinPercent)
			winMultiplier := multiplier * (1 - HouseEdge)
			result = settleDiceResult(input.WinPercent, input.Direction)

			if result.Won {
				toCashBalance = fromCashBalance * winMultiplier
				toBonusBalance = fromBonusBalance * winMultiplier

				win := models.Winning{
					UserID:    user.ID,
					WinAmount: toCashBalance + toBonusBalance,
				}
			
				if err := tx.Create(&win).Error; err != nil {
					return logger.WrapError(err, "Failed to record winning")
				}
			}

			// Update both balances even if won is false
			err = exchange.UpdateUserBalances(tx, &user, toCashBalance, toBonusBalance, false)
			if err != nil {
				return logger.WrapError(err, "")
			}
		} else {
			// If there are free bet available
			// Calculate winnings and settle result
			multiplier := 100 / float64(input.WinPercent)
			winMultiplier := multiplier * (1 - HouseEdge)
			result = settleDiceResult(input.WinPercent, input.Direction)

			if result.Won {
				toBonusBalance = benefitFreeDeposit * winMultiplier
				

				win := models.Winning{
					UserID:    user.ID,
					WinAmount: toBonusBalance,
				}

				if err := tx.Create(&win).Error; err != nil {
					return logger.WrapError(err, "Failed to record winning")
				}

				// Update both balances
				err = exchange.UpdateUserBalances(tx, &user, 0, toBonusBalance, true)
				if err != nil {
					return logger.WrapError(err, "")
				}
			}

			if err = applyBenefit(tx); err != nil {
				return logger.WrapError(err, "")
			}
		}

		if benefitFreeDeposit == 0 {
			if err = updateDiceGameTravePassLevelRequirement(tx, user.ID, fromCashBalance, toCashBalance); err != nil {
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

	c.JSON(200, gin.H{
		"result": result,
		"payout": toCashBalance + toBonusBalance,
	})
}

func updateDiceGameTravePassLevelRequirement(
	tx *gorm.DB, userID int64, fromCashBalance, toCashBalance float64) error {
	if tx == nil {
		tx = db.DB
	}

	requirementMGDone, err := requirement_progress.UpdateRequirementProgressMiniGameIfRequired(
		tx, userID, requirements.DiceGameID, fromCashBalance, toCashBalance)
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementTurnoverDone, err := requirement_progress.UpdateRequirementProgressTurnoverIfRequired(
		tx, userID, fromCashBalance)
	if err != nil {
		return logger.WrapError(err, "")
	}

	if requirementMGDone || requirementTurnoverDone {
		if err = travepass.CheckAndUpgradeTravePassLevel(tx, userID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}

func settleDiceResult(winPercent int64, direction DiceBetDirection) DiceBetResult {
	number := rand.Intn(1000000)
	threshold := int(winPercent) * 10000

	var won bool
	if direction == DiceBetLess {
		won = number < threshold
	} else {
		won = number > (999999 - threshold)
	}

	return DiceBetResult{
		Won:    won,
		Number: number,
	}
}
