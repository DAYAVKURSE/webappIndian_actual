package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/logger"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ReferrerFirstDepositBonusPercentage = 0.2

func GetUserDeposits(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var deps []models.Deposit
	err = db.DB.Find(&deps, "user_id = ?", userID).Error
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	if len(deps) == 0 {
		c.String(404, "[]")
		return
	}

	c.JSON(200, deps)
}

// AddDeposit adds deposit from transaction to user.
// If error occures, function will set corresponding status to gin
// and return false. If there are no error, function will return true
func AddDeposit(c *gin.Context, transactionBody TransactionBody) (success bool) {
	var err error
	var dep models.Deposit
	bytes, _ := json.Marshal(transactionBody)

	var user models.User
	user.ID, err = strconv.ParseInt(*transactionBody.CustomUserID, 10, 64)
	if err != nil {
		logger.Error("Unable to convert transaction custom_user_id to int\n%s\n%v", string(bytes), err)
		c.JSON(400, gin.H{"error": "unable to convert transaction custom user id to int"})
		return false
	}

	exists, err := models.CheckIfUserExistsByID(user.ID)
	if err != nil {
		logger.Error("Unable to check if user with this id exists\n%s\n%v", string(bytes), err)
		c.Status(500)
		return false
	}

	if !exists {
		logger.Error("Incoming deposit for non-existing user\n%s", string(bytes))
		c.JSON(404, gin.H{"error": "user not found"})
		return false
	}

	// check if deposit with this orderID exists
	err = db.DB.Model(&models.Deposit{}).
		Select("count(*) > 0").
		Where("order_id = ?", transactionBody.OrderID).
		Scan(&exists).Error
	if err != nil {
		logger.Error("Unable to check if deposit with orderID exists\n%s\n%v", string(bytes), err)
		c.Status(500)
		return false
	}

	if exists {
		logger.Error("Deposit with this order id already exists\n%s", string(bytes))
		c.JSON(409, gin.H{"error": "deposit with this order id already exists"})
		return false
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Lock user row for update to prevent race conditions
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(
			&user, user.ID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		bonusMultiplier, err := benefit_progress.GetBenefitReplenishmentProgressBonusMultiplier(tx, user.ID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		// Create the deposit
		dep = models.Deposit{
			UserID:         user.ID,
			AmountRupee:    transactionBody.Amount * bonusMultiplier,
			BonusExpiresIn: time.Now().Add(models.DepositBonusDuration),
			OrderID:        transactionBody.OrderID,
		}

		if err = tx.Create(&dep).Error; err != nil {
			return logger.WrapError(err, "")
		}

		// Update user's balance
		if err = tx.Model(&user).Update("balance_rupee", gorm.Expr(
			"balance_rupee + ?", transactionBody.Amount*bonusMultiplier)).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if err = updateReplinishmentTravePassLevelRequirements(tx, user.ID, transactionBody.Amount); err != nil {
			return logger.WrapError(err, "")
		}

		if err = giveRefereeBonus(tx, &dep, transactionBody.Amount); err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil {
		logger.Error("Failed to process deposit\n%s\n%v", string(bytes), err)
		c.Status(500)
		return false
	}

	return true
}

func updateReplinishmentTravePassLevelRequirements(
	tx *gorm.DB, userID int64, amountRupee float64) error {
	if tx == nil {
		tx = db.DB
	}

	requirementReplenishmentDone, err := requirement_progress.UpdateRequirementProgressReplenishmentIfRequired(
		tx, userID, amountRupee)
	if err != nil {
		return logger.WrapError(err, "")
	}

	if requirementReplenishmentDone {
		if err = travepass.CheckAndUpgradeTravePassLevel(tx, userID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}

func giveRefereeBonus(tx *gorm.DB, dep *models.Deposit, rupeeAmountWithoutMultiplier float64) error {
	var otherDepsExists, userIsAReferral bool

	// Check that created deposit is user's first
	err := tx.Model(&models.Deposit{}).
		Select("count(*) > 1").
		Where("user_id = ?", dep.UserID).
		Scan(&otherDepsExists).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	// Check that user is a referral
	err = tx.Model(&models.UserReferral{}).
		Select("count(*) > 0").
		Where("referred_id = ?", dep.UserID).
		Scan(&userIsAReferral).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	if !otherDepsExists && userIsAReferral {
		var userRef models.UserReferral
		if err = tx.Where("referred_id = ?", dep.UserID).First(&userRef).Error; err != nil {
			return logger.WrapError(err, "")
		}

		referrerBonusToAdd := rupeeAmountWithoutMultiplier * ReferrerFirstDepositBonusPercentage

		// Update referrer's balance
		if err = tx.Model(&models.User{}).
			Where("id = ?", userRef.ReferrerID).
			Update("balance_rupee", gorm.Expr(
				"balance_rupee + ?", referrerBonusToAdd)).Error; err != nil {
			return logger.WrapError(err, "")
		}

		userRef.ReferredFirstDepositID = &dep.ID
		userRef.EarnedAmount += referrerBonusToAdd

		if err = tx.Save(&userRef).Error; err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}
