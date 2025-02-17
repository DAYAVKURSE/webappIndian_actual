package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/logger"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type exchangeInput struct {
	AmountBcoins float64 `validate:"required,min=1"`
}

func (i *exchangeInput) Validate() error {
	validate := validator.New()
	return validate.Struct(i)
}

func ExchangeBcoinsToRupee(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var input exchangeInput
	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Unable to unmarshal body"})
		return
	}

	if err = input.Validate(); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User

		// Lock user row for update to prevent race conditions
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if user.BalanceBi < input.AmountBcoins {
			c.JSON(402, gin.H{"error": "Insufficient balance"})
			return nil
		}

		user.BalanceBi -= input.AmountBcoins
		if err := tx.Save(&user).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if err = exchange.CreateOrUpdateExchangeBalance(
			tx, userID, input.AmountBcoins*exchange.BCoinsToRupeeCourseMultiplier); err != nil {
			return logger.WrapError(err, "")
		}

		if err = updateExchangeTravePassLevelRequirements(
			tx, userID, input.AmountBcoins); err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil {
		logger.Error("%v", err)
		return
	}

	c.Status(200)
}

func updateExchangeTravePassLevelRequirements(
	tx *gorm.DB, userID int64, amountBcoins float64) error {
	if tx == nil {
		tx = db.DB
	}

	requirementExchangeDone, err := requirement_progress.
		UpdateRequirementProgressExchangeIfRequired(tx, userID, amountBcoins)
	if err != nil {
		return logger.WrapError(err, "")
	}

	if requirementExchangeDone {
		if err = travepass.CheckAndUpgradeTravePassLevel(tx, userID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}

func GetUserExchangeBalance(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	errNotFound := errors.New("user ExchangeBalance not found")

	var exchangeBalance exchange.ExchangeBalance
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.First(&exchangeBalance, "user_id = ?", userID).Error
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			return errNotFound
		} else if err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil && errors.Is(err, errNotFound) {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.JSON(200, exchangeBalance)
}
