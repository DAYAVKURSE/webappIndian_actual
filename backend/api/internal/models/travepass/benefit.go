package travepass

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/pkg/logger"

	"gorm.io/gorm"
)

type TravePassLevelBenefit struct {
	ID               int64            `gorm:"primaryKey;autoIncrement"`
	TravePassLevelID int64            `gorm:"index"`
	BenefitID        int64            `gorm:"index"`
	Benefit          benefits.Benefit `gorm:"foreignKey:BenefitID;constraint:OnDelete:SET NULL;"`
}

// CreateUserBenefitProgresses creates BenefitProgresses with trave pass level id.
// Should be used on trave pass level up.
func CreateUserBenefitProgresses(tx *gorm.DB, userID, currentLevelID int64) error {
	if tx == nil {
		tx = db.DB
	}

	var levelBenefits []TravePassLevelBenefit

	err := tx.Preload("Benefit").Find(&levelBenefits, "trave_pass_level_id = ?", currentLevelID).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	for reqIdx := range levelBenefits {
		if err = levelBenefits[reqIdx].Benefit.PreloadPolymorphicBenefit(tx); err != nil {
			return logger.WrapError(err, "")
		}
		if err = benefit_progress.CreateOrApplyPolymorphicBenefitProgress(
			tx, &levelBenefits[reqIdx].Benefit, userID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}
