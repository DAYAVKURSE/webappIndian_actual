package exchange

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

const (
	WagerMultiplier               = 5
	BCoinsToRupeeCourseMultiplier = 0.0001
)

// ExchangeBalance is designed to implement a 5x wagering requirement for
// withdrawing funds earned without spending real money. If a user has
// a bonus balance, bets are deducted from it first, and if it's insufficient,
// the user's main balance is used. When a bet wins, both balances are updated
// proportionally to the amount spent from each. Once the required turnover is
// reached, the bonus balance is transferred to the main balance, allowing the
// user to withdraw winnings. If the bonus balance reaches zero after a bonus-funded
// bet, it is fully removed. Turnover made with the bonus balance does not count
// toward completing `trave pass` tasks, and real money turnover doesn't reduce
// the bonus wagering requirement. All winnings from free bets (e.g., from a
// wheel of fortune or a pass) are added to the bonus balance, increasing the
// required turnover for withdrawal.
type ExchangeBalance struct {
	ID               int64 `gorm:"primaryKey;autoIncrement"`
	UserID           int64 `gorm:"index"`
	AmountRupee      float64
	CurrentTurnover  float64
	RequiredTurnover float64
}

// CreateOrUpdateExchangeBalance checks if user already have ExchangeBalance
// and if it exists, adds bonus money on it and increase wager required turnover.
// If user dont have ExchangeBalance, it will be created with amountRupee parameter.
func CreateOrUpdateExchangeBalance(tx *gorm.DB, userID int64, amountRupee float64) error {
	if tx == nil {
		tx = db.DB
	}

	var existingExchangeBalance ExchangeBalance
	err := tx.First(&existingExchangeBalance, "user_id = ?", userID).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		existingExchangeBalance = ExchangeBalance{}
	} else if err != nil {
		return logger.WrapError(err, "")
	}

	existingExchangeBalance.UserID = userID
	existingExchangeBalance.AmountRupee += amountRupee
	existingExchangeBalance.RequiredTurnover += amountRupee * WagerMultiplier

	if err = tx.Save(&existingExchangeBalance).Error; err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

func GetUserExchangedBalanceAmount(tx *gorm.DB, userID int64) (float64, error) {
	if tx == nil {
		tx = db.DB
	}

	var exchangedBalanceAmount sql.NullFloat64
	err := tx.Model(&ExchangeBalance{}).Where("user_id = ?", userID).
		Pluck("amount_rupee", &exchangedBalanceAmount).Error
	if err != nil {
		return 0, logger.WrapError(err, "")
	}

	if exchangedBalanceAmount.Valid {
		return exchangedBalanceAmount.Float64, nil
	}

	return 0, nil
}

// UseExchangeBalancePayment should be called on each in-application payment.
// It checks if user have ExchangeBalance and determines amount of money, that
// should be spent from both user's balances. If user dont have ExchangeBalance
// money will be spend from real user balance.
func UseExchangeBalancePayment(tx *gorm.DB, user *models.User, amountRupee float64) (
	fromCashBalance float64, fromBonusBalance float64, err error) {
	if tx == nil {
		tx = db.DB
	}

	var exchangeBalance ExchangeBalance
	err = tx.First(&exchangeBalance, "user_id = ?", user.ID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, logger.WrapError(err, "")
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Exchange balance not exists, use user balance
		fromCashBalance = amountRupee
	} else {
		if user.BalanceRupee+exchangeBalance.AmountRupee < amountRupee {
			return 0, 0, logger.WrapError(errors.New("insufficient balance"), "")
		}

		// Use exchange balance first
		if exchangeBalance.AmountRupee < amountRupee {
			fromBonusBalance = exchangeBalance.AmountRupee
			fromCashBalance = amountRupee - exchangeBalance.AmountRupee
			exchangeBalance.AmountRupee = 0
		} else {
			exchangeBalance.AmountRupee -= amountRupee
			fromBonusBalance = amountRupee
		}

		exchangeBalance.CurrentTurnover += fromBonusBalance
		if err = tx.Save(&exchangeBalance).Error; err != nil {
			return 0, 0, logger.WrapError(err, "")
		}
	}

	user.BalanceRupee -= fromCashBalance
	if err = tx.Save(user).Error; err != nil {
		return 0, 0, logger.WrapError(err, "")
	}

	return fromCashBalance, fromBonusBalance, nil
}

// UpdateUserBalances should be called straight after the bet result is clear
// regardless, result is won or not and user ExchangeBalance exists or not.
// Function attempts to fetch the user's ExchangeBalance record, which manages
// bonus balance and its turnover requirement.- If the record does not exist
// or a benefit win occurred, and if a bonus amount (toBonusBalance) is provided,
// the function creates or updates the ExchangeBalance for the user.
// If the ExchangeBalance exists, it updates the bonus amount and checks whether
// the required turnover has been met. - If the turnover is complete or the bonus
// balance is zero, the remaining bonus balance is moved to the user's cash balance
// and the ExchangeBalance record is deleted. - If the turnover is incomplete, the
// ExchangeBalance is updated and saved. The user's cash balance (toCashBalance) is
// always updated and saved, regardless of the ExchangeBalance status.
func UpdateUserBalances(tx *gorm.DB, user *models.User, toCashBalance,
	toBonusBalance float64, benefitWin bool) error {
	if tx == nil {
		tx = db.DB
	}

	var exchangeBalance ExchangeBalance
	err := tx.First(&exchangeBalance, "user_id = ?", user.ID).Error

	// ExchangeBalance not exists or requiredTurnover should be updated
	if (err != nil && errors.Is(err, gorm.ErrRecordNotFound)) || benefitWin {
		// If there are bonusBalance to add, ExchangeBalance should be created
		if toBonusBalance != 0 {
			if err = CreateOrUpdateExchangeBalance(tx, user.ID, toBonusBalance); err != nil {
				return logger.WrapError(err, "")
			}
		}
	} else if err != nil {
		return logger.WrapError(err, "")
	} else {
		// ExchangeBalance exists
		exchangeBalance.AmountRupee += toBonusBalance

		// ExchangeBalance requiredTuronver reached or ExchangeBalance AmountRupee still equals zero
		if exchangeBalance.AmountRupee == 0 || exchangeBalance.CurrentTurnover >= exchangeBalance.RequiredTurnover {
			user.BalanceRupee += exchangeBalance.AmountRupee
			if err := tx.Delete(&exchangeBalance).Error; err != nil {
				return logger.WrapError(err, "")
			}
		} else {
			// Otherwise, save the updated exchange balance
			if err := tx.Save(&exchangeBalance).Error; err != nil {
				return logger.WrapError(err, "")
			}
		}
	}

	// Add cashBalance to user and save it,
	// regardless of whether ExchangeBalance exists or not
	user.BalanceRupee += toCashBalance
	if err := tx.Save(&user).Error; err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}
