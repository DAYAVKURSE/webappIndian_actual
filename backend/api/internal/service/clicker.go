package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/logger"
	"errors"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type clicksInput struct {
	ClicksCount int `validate:"required,min=0,max=1000"`
	BiPerClick  float64
}

func (i *clicksInput) Validate() error {
	validate = validator.New()
	return validate.Struct(i)
}

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func AddClicks(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var input clicksInput
	if err = c.Bind(&input); err != nil {
		c.JSON(400, gin.H{"error": "unable to unmarshal body"})
		return
	}

	err = input.Validate()
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	calculatedBiPerClick, err := models.CountUserBiPerClick(userID)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Compare input BiPerClick with calculated value
	biPerClickMismatch := !almostEqual(input.BiPerClick, calculatedBiPerClick)

	// Always use the calculated BiPerClick for balance updates
	biPerClickToUse := calculatedBiPerClick

	errClicksCountOutOfBorders := errors.New("clicks count out of borders")
	var addedClicksCount int
	var bonusMultiplier float64

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		err = tx.First(&user, userID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}

		if user.DailyClicks >= models.DailyClicksLimit {
			return errClicksCountOutOfBorders
		}

		bonusMultiplier, err = benefit_progress.GetBenefitClickerProgressBonusMultiplier(tx, user.ID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		// Update user's clicks and balance
		if user.DailyClicks+input.ClicksCount <= models.DailyClicksLimit {
			user.DailyClicks += input.ClicksCount
			user.BalanceBi += float64(input.ClicksCount) * biPerClickToUse * bonusMultiplier
			addedClicksCount = input.ClicksCount
		} else {
			clicksToAdd := models.DailyClicksLimit - user.DailyClicks
			user.DailyClicks = models.DailyClicksLimit
			user.BalanceBi += float64(clicksToAdd) * biPerClickToUse * bonusMultiplier
			addedClicksCount = clicksToAdd
		}

		err = db.DB.Save(&user).Error
		if err != nil {
			return logger.WrapError(err, "")
		}

		if err = updateClickerTravePassLevelRequirements(tx, &user, addedClicksCount); err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil && errors.Is(err, errClicksCountOutOfBorders) {
		c.JSON(417, gin.H{"error": err.Error()}) // http.NotAcceptable
		return
	} else if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Respond based on whether there was a mismatch
	if biPerClickMismatch {
		c.JSON(200, gin.H{
			"BiPerClick":      calculatedBiPerClick,
			"BonusMultiplier": bonusMultiplier,
		})
	} else {
		c.Status(200)
	}
}

func updateClickerTravePassLevelRequirements(
	tx *gorm.DB, user *models.User, addedClicksCount int) error {
	if tx == nil {
		tx = db.DB
	}

	requirementClickerDone, err := requirement_progress.UpdateRequirementProgressClickerIfRequired(tx, user.ID, addedClicksCount)
	if err != nil {
		return logger.WrapError(err, "")
	}

	if requirementClickerDone {
		if err = travepass.CheckAndUpgradeTravePassLevel(tx, user.ID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}

func GetUserCurrentBiPerClickCost(c *gin.Context) {
	// get user id
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// get click cost
	biPerClick, err := models.CountUserBiPerClick(userID)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	bonusMultiplier, err := benefit_progress.GetBenefitClickerProgressBonusMultiplier(nil, userID)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// return click cost
	c.JSON(200, gin.H{
		"BiPerClick":      biPerClick,
		"BonusMultiplier": bonusMultiplier,
	})
}
