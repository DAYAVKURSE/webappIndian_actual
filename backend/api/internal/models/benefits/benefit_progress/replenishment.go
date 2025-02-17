package benefit_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/pkg/logger"
	"errors"
	"time"

	"gorm.io/gorm"
)

const BenefitProgressReplenishmentType = "benefit_replenishment_progress"

type BenefitProgressReplenishment struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	ValidUntil      time.Time
	BonusMultiplier float64
}

// CreateBenefitProgressReplenishment creates BenefitProgressReplenishment and
// linked BenefitProgress. Benefit parameter should contain existing
// polymorphic benefit.
func CreateBenefitProgressReplenishment(tx *gorm.DB, benefit *benefits.Benefit, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	benefitReplenishment, ok := benefit.PolymorphicBenefit.(benefits.BenefitReplenishment)
	if !ok {
		return logger.WrapError(errors.New(
			"unable to cast benefit.PolymorphicBenefit to BenefitReplenishment"), "")
	}

	benefitProgressReplenishment := BenefitProgressReplenishment{
		ValidUntil: time.Now().Add(
			time.Duration(benefitReplenishment.TimeDuration) * time.Second),
		BonusMultiplier: benefitReplenishment.BonusMultiplier,
	}

	err := tx.Save(&benefitProgressReplenishment).
		Scan(&benefitProgressReplenishment).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	err = tx.Save(&BenefitProgress{
		UserID:                         userID,
		BenefitID:                      benefit.ID,
		PolymorphicBenefitProgressID:   benefitProgressReplenishment.ID,
		PolymorphicBenefitProgressType: BenefitProgressReplenishmentType}).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// GetBenefitReplenishmentProgressBonusMultiplier checks if there are available
// replenishment bonus and returns bcoins multiplier. If ValidUntil reached,
// BenefitProgressReplenishment with BenefitProgress will be deleted.
func GetBenefitReplenishmentProgressBonusMultiplier(tx *gorm.DB, UserID int64) (bonusMultiplier float64, err error) {
	if tx == nil {
		tx = db.DB
	}

	bonusMultiplier = 1

	var benefitProgresses []BenefitProgress
	err = tx.Find(&benefitProgresses, "user_id = ? and polymorphic_benefit_progress_type = ?",
		UserID, BenefitProgressReplenishmentType).Error
	if err != nil {
		return 1, logger.WrapError(err, "")
	}

	if len(benefitProgresses) == 0 {
		return 1, nil
	}

	for i := range benefitProgresses {
		err = benefitProgresses[i].PreloadPolymorphicBenefitProgress(tx)
		if err != nil {
			return 1, logger.WrapError(err, "")
		}

		benefitProgressReplenishment, ok := benefitProgresses[i].PolymorphicBenefitProgress.(BenefitProgressReplenishment)
		if !ok {
			return 1, logger.WrapError(errors.New(
				"unable to cast PolymorphicBenefitProgress to BenefitProgressReplenishment"), "")
		}

		// benefit expired
		if time.Now().After(benefitProgressReplenishment.ValidUntil) {
			if err = tx.Delete(benefitProgresses[i]).Error; err != nil {
				return 1, logger.WrapError(err, "")
			}

			if err = tx.Delete(benefitProgressReplenishment).Error; err != nil {
				return 1, logger.WrapError(err, "")
			}
		} else {
			// stack bonus multipliers
			bonusMultiplier += benefitProgressReplenishment.BonusMultiplier
		}
	}

	return bonusMultiplier, nil
}
