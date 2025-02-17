package benefit_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const BenefitProgressFortuneWheelType = "benefit_fortune_wheel_progress"

type BenefitProgressFortuneWheel struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	FreeSpinsAmount int
}

// CreateBenefitProgressFortuneWheel creates BenefitProgressFortuneWheel and
// linked BenefitProgress. Benefit parameter should contain existing
// polymorphic benefit.
func CreateBenefitProgressFortuneWheel(tx *gorm.DB, benefit *benefits.Benefit, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	benefitFortuneWheel, ok := benefit.PolymorphicBenefit.(benefits.BenefitFortuneWheel)
	if !ok {
		return logger.WrapError(errors.New(
			"unable to cast benefit.PolymorphicBenefit to BenefitFortuneWheel"), "")
	}

	benefitProgressFortuneWheel := BenefitProgressFortuneWheel{
		FreeSpinsAmount: benefitFortuneWheel.FreeSpinsAmount,
	}

	err := tx.Save(&benefitProgressFortuneWheel).
		Scan(&benefitProgressFortuneWheel).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	err = tx.Save(&BenefitProgress{
		UserID:                         userID,
		BenefitID:                      benefit.ID,
		PolymorphicBenefitProgressID:   benefitProgressFortuneWheel.ID,
		PolymorphicBenefitProgressType: BenefitProgressFortuneWheelType}).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UseFreeFortuneWheelSpinIfAvailable checks if there are available benefits
// on fortune wheel for user, and returns true if spin is available
// with function to update free spins count. If last free spin used,
// BenefitProgressFortuneWheel with BenefitProgress will be deleted.
func UseFreeFortuneWheelSpinIfAvailable(tx *gorm.DB, userID int64) (
	bool, func(tx *gorm.DB) error, error) {
	if tx == nil {
		tx = db.DB
	}

	var benefitProgress BenefitProgress
	err := tx.First(&benefitProgress,
		"user_id = ? and polymorphic_benefit_progress_type = ?",
		userID, BenefitProgressFortuneWheelType).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// User dont have free spins
		return false, nil, nil
	} else if err != nil {
		return false, nil, logger.WrapError(err, "")
	}

	if err = benefitProgress.PreloadPolymorphicBenefitProgress(tx); err != nil {
		return false, nil, logger.WrapError(err, "")
	}

	benefitProgressFortuneWheel, ok := benefitProgress.
		PolymorphicBenefitProgress.(BenefitProgressFortuneWheel)
	if !ok {
		return false, nil, logger.WrapError(errors.New(
			"unable to cast PolymorphicBenefitProgress to BenefitProgressFortuneWheel"), "")
	}

	return true, func(tx *gorm.DB) error {
		benefitProgressFortuneWheel.FreeSpinsAmount--
		if benefitProgressFortuneWheel.FreeSpinsAmount == 0 {
			err = tx.Delete(&benefitProgressFortuneWheel).Error
			if err != nil {
				return logger.WrapError(err, "")
			}

			err = tx.Delete(&benefitProgress).Error
			if err != nil {
				return logger.WrapError(err, "")
			}
		} else if err = tx.Save(&benefitProgressFortuneWheel).Error; err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	}, nil
}

func GetUserAvailableSpins(tx *gorm.DB, userID int64) (int, error) {
	if tx == nil {
		tx = db.DB
	}

	var benefitProgresses []BenefitProgress
	err := tx.Find(&benefitProgresses,
		"user_id = ? and polymorphic_benefit_progress_type = ?",
		userID, BenefitProgressFortuneWheelType).Error
	if err != nil {
		return 0, logger.WrapError(err, "")
	}

	if len(benefitProgresses) == 0 {
		return 0, nil
	}

	totalSpins := 0
	for i := range benefitProgresses {
		if err = benefitProgresses[i].PreloadPolymorphicBenefitProgress(tx); err != nil {
			return 0, logger.WrapError(err, "")
		}
		totalSpins += benefitProgresses[i].PolymorphicBenefitProgress.(BenefitProgressFortuneWheel).FreeSpinsAmount
	}

	return totalSpins, nil
}
