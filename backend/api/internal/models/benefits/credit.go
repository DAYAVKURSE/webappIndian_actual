package benefits

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"

	"gorm.io/gorm"
)

const BenefitCreditType = "benefit_credit"

type BenefitCredit struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	BCoinsAmount float64
	RupeeAmount  float64
}

func (benefitCredit *BenefitCredit) ApplyBenefit(tx *gorm.DB, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"balance_rupee": gorm.Expr("balance_rupee + ?", benefitCredit.RupeeAmount),
		"balance_bi":    gorm.Expr("balance_bi + ?", benefitCredit.BCoinsAmount),
	}).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}
