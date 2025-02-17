package models

import (
	"BlessedApi/cmd/db"
	"BlessedApi/pkg/logger"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

var DepositBonusDuration = time.Hour * 24 * 20
var BiRupeeCourse = 10000
var MinDepositRupee = 500

type Deposit struct {
	ID             int64 `gorm:"primaryKey,autoIncrement"`
	UserID         int64 `gorm:"index;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	OrderID        string
	AmountRupee    float64
	BonusExpiresIn time.Time
	CreatedAt      time.Time
}

func GetUserTotalDeposit(tx *gorm.DB, userID int64) (float64, error) {
	if tx == nil {
		tx = db.DB
	}

	var sum sql.NullFloat64
	if err := tx.Model(&Deposit{}).
		Where("user_id = ?", userID).
		Select("SUM(amount_rupee)").
		Scan(&sum).Error; err != nil {
		// user should exists when this function will be available to call
		return 0, logger.WrapError(err, "")
	}

	if sum.Valid {
		return sum.Float64, nil
	}

	return 0, nil
}
