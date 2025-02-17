package models

import (
	"BlessedApi/cmd/db"
	"BlessedApi/pkg/logger"
	"time"

	"gorm.io/gorm"
)

type BinaryBetDirection string

const (
	BetUp   BinaryBetDirection = "up"
	BetDown BinaryBetDirection = "down"
)

type BinaryBet struct {
	ID               int64   `gorm:"primaryKey,autoIncrement"`
	UserID           int64   `gorm:"not null;index"`
	Amount           float64 `gorm:"not null"`
	FromBonusBalance float64 `json:"-"`
	FromCashBalance  float64 `json:"-"`
	IsBenefitBet     bool
	Direction        BinaryBetDirection `gorm:"not null"`
	OpenedAt         time.Time          `gorm:"not null"`
	ExpiresAt        time.Time          `gorm:"not null"`
	OpenPrice        float64            `gorm:"not null"`
	ClosePrice       float64
	Outcome          string
	Payout           float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func UserHasActiveBet(tx *gorm.DB, userID int64) (bool, error) {
	if tx == nil {
		tx = db.DB
	}

	var count int64
	err := tx.Model(&BinaryBet{}).
		Where("user_id = ? AND expires_at > ?",
			userID, time.Now()).
		Count(&count).
		Error
	if err != nil {
		return false, logger.WrapError(err, "")
	}
	return count > 0, nil
}

func MaintainLastTenBinaryOptionBets(tx *gorm.DB, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	var count int64
	if err := tx.Model(&BinaryBet{}).
		Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return logger.WrapError(err, "")
	}

	if count > 10 {
		var oldestBets []BinaryBet
		if err := tx.Where("user_id = ?", userID).
			Order("created_at asc").Limit(int(count - 10)).
			Find(&oldestBets).Error; err != nil {
			return logger.WrapError(err, "")
		}

		for _, bet := range oldestBets {
			if err := tx.Delete(&bet).Error; err != nil {
				return logger.WrapError(err, "")
			}
		}
	}

	return nil
}
