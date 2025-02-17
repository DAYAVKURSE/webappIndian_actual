package benefit_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/pkg/logger"
	"errors"
	"time"

	"gorm.io/gorm"
)

const BenefitProgressClickerType = "benefit_clicker_progress"

type BenefitProgressClicker struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	ValidUntil      time.Time
	BonusMultiplier float64
}

// CreateOrApplyBenefitProgressClicker creates BenefitProgressClicker and
// linked BenefitProgress. Benefit parameter should contain existing
// polymorphic benefit.
func CreateOrApplyBenefitProgressClicker(tx *gorm.DB, benefit *benefits.Benefit, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	benefitClicker, ok := benefit.PolymorphicBenefit.(benefits.BenefitClicker)
	if !ok {
		return logger.WrapError(errors.New(
			"unable to cast benefit.PolymorphicBenefit to BenefitClicker"), "")
	}

	if benefitClicker.Reset {
		err := tx.Model(&models.User{}).Where("id = ?", userID).Update("daily_clicks", 0).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		return nil
	}

	benefitProgressClicker := BenefitProgressClicker{
		ValidUntil: time.Now().Add(
			time.Duration(benefitClicker.TimeDuration) * time.Second),
		BonusMultiplier: benefitClicker.BonusMultiplier,
	}

	err := tx.Save(&benefitProgressClicker).
		Scan(&benefitProgressClicker).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	err = tx.Save(&BenefitProgress{
		UserID:                         userID,
		BenefitID:                      benefit.ID,
		PolymorphicBenefitProgressID:   benefitProgressClicker.ID,
		PolymorphicBenefitProgressType: BenefitProgressClickerType}).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// GetBenefitClickerProgressBonusMultiplier checks if there are available
// clicker bonus and returns bcoins multiplier. If ValidUntil time reached,
// BenefitProgressClicker with BenefitProgress will be deleted.
func GetBenefitClickerProgressBonusMultiplier(tx *gorm.DB, UserID int64) (
	bonusMultiplier float64, err error) {
	if tx == nil {
		tx = db.DB
	}

	var benefitProgresses []BenefitProgress
	err = tx.Find(&benefitProgresses,
		"user_id = ? and polymorphic_benefit_progress_type = ?",
		UserID, BenefitProgressClickerType).Error
	if err != nil {
		return 1, logger.WrapError(err, "")
	}

	if len(benefitProgresses) == 0 {
		return 1, nil
	}

	bonusMultiplier = 0
	for i := range benefitProgresses {
		err = benefitProgresses[i].PreloadPolymorphicBenefitProgress(tx)
		if err != nil {
			return 1, logger.WrapError(err, "")
		}

		benefitProgressClicker, ok := benefitProgresses[i].PolymorphicBenefitProgress.(BenefitProgressClicker)
		if !ok {
			return 1, logger.WrapError(errors.New(
				"unable to cast PolymorphicBenefitProgress to BenefitProgressClicker"), "")
		}

		// benefit expired
		if time.Now().After(benefitProgressClicker.ValidUntil) {
			if err = tx.Delete(benefitProgresses[i]).Error; err != nil {
				return 1, logger.WrapError(err, "")
			}

			if err = tx.Delete(benefitProgressClicker).Error; err != nil {
				return 1, logger.WrapError(err, "")
			}
		} else {
			// stack bonus multipliers
			bonusMultiplier += benefitProgressClicker.BonusMultiplier
		}
	}

	return bonusMultiplier, nil
}
