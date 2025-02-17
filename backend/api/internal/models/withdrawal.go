package models

import (
	"BlessedApi/cmd/db"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const MinDepositSumToWithdrawal = 13000

type Withdrawal struct {
	ID        int64 `gorm:"primaryKey,autoIncrement"`
	UserID    int64 `gorm:"index;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Amount    float64
	OrderID   string
	CreatedAt int64
	Status    string
}

func (w *Withdrawal) Rollback() error {
	if w.UserID == 0 || w.Amount == 0 {
		return nil
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).
			Where("id = ?", w.UserID).
			Update("balance_rupee",
				gorm.Expr("balance_rupee + ?", w.Amount)).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if err := tx.Delete(w).Error; err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil {
		return logger.WrapError(err, "")
	}

	logger.Debug("Withdrawal rollbacked. Order id: %s", w.OrderID)

	return nil
}

// Checks by order id if there are withdrawal
// and sets new status with True return value. If there are no withdrawal record
// returns False.
func UpdateWithdrawalStatusIfRequired(orderID, status string) (bool, error) {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var withdrawal Withdrawal
		if err := tx.First(&withdrawal, "order_id = ?", orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return logger.WrapError(err, "")
		}

		if status != "Success" {
			if err := withdrawal.Rollback(); err != nil {
				return logger.WrapError(err, "")
			}
			return nil
		}

		withdrawal.Status = "Success"
		if err := tx.Save(&withdrawal).Error; err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		return false, logger.WrapError(err, "")
	}

	return true, nil
}

