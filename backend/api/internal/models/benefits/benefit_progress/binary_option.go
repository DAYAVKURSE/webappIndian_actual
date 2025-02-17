package benefit_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const BenefitProgressBinaryOptionType = "benefit_binary_option_progress"

type BenefitProgressBinaryOption struct {
	ID                  int64 `gorm:"primaryKey;autoIncrement"`
	FreeBetsAmount      int
	FreeBetDepositRupee float64
}

// CreateBenefitProgressBinaryOption creates BenefitProgressBinaryOption and
// linked BenefitProgress. Benefit parameter should contain existing
// polymorphic benefit.
func CreateBenefitProgressBinaryOption(tx *gorm.DB, benefit *benefits.Benefit, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	benefitBinaryOpt, ok := benefit.PolymorphicBenefit.(benefits.BenefitBinaryOption)
	if !ok {
		return logger.WrapError(errors.New(
			"unable to cast benefit.PolymorphicBenefit to BenefitClicker"), "")
	}

	benefitProgressBinaryOption := BenefitProgressBinaryOption{
		FreeBetsAmount:      benefitBinaryOpt.FreeBetsAmount,
		FreeBetDepositRupee: benefitBinaryOpt.FreeBetDepositRupee,
	}

	err := tx.Save(&benefitProgressBinaryOption).
		Scan(&benefitProgressBinaryOption).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	err = tx.Save(&BenefitProgress{
		UserID:                         userID,
		BenefitID:                      benefit.ID,
		PolymorphicBenefitProgressID:   benefitProgressBinaryOption.ID,
		PolymorphicBenefitProgressType: BenefitProgressBinaryOptionType}).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UseFreeBinaryOptionBetIfAvailable checks if there are available benefits
// on binary option for user, and returns free bet deposit with function to
// update free bets count. If the last free bet used, BenefitProgressBinaryOption
// with BenefitProgress will be deleted.
func UseFreeBinaryOptionBetIfAvailable(tx *gorm.DB, userID int64) (
	float64, func(tx *gorm.DB) error, error) {
	if tx == nil {
		tx = db.DB
	}

	var benefitProgress BenefitProgress
	err := tx.First(&benefitProgress,
		"user_id = ? and polymorphic_benefit_progress_type = ?",
		userID, BenefitProgressBinaryOptionType).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// no available benefits
		return 0, nil, nil
	} else if err != nil {
		return 0, nil, logger.WrapError(err, "")
	}

	if err = benefitProgress.PreloadPolymorphicBenefitProgress(tx); err != nil {
		return 0, nil, logger.WrapError(err, "")
	}

	benefitProgressBinaryOption, ok := benefitProgress.
		PolymorphicBenefitProgress.(BenefitProgressBinaryOption)
	if !ok {
		return 0, nil, logger.WrapError(errors.New(
			"unable to cast PolymorphicBenefitProgress to BenefitProgressBinaryOption"), "")
	}

	return benefitProgressBinaryOption.FreeBetDepositRupee, func(tx *gorm.DB) error {
		benefitProgressBinaryOption.FreeBetsAmount--

		if benefitProgressBinaryOption.FreeBetsAmount == 0 {
			if err = tx.Delete(&benefitProgressBinaryOption).Error; err != nil {
				return logger.WrapError(err, "")
			}

			if err = tx.Delete(&benefitProgress).Error; err != nil {
				return logger.WrapError(err, "")
			}
		} else if err = tx.Save(&benefitProgressBinaryOption).Error; err != nil {
			return logger.WrapError(err, "")
		}
		return nil
	}, nil
}

// GetUserFreeBinaryOptionBets returns user available free binary option bets.
// Function returns pointer to array of BenefitProgressBinaryOption to give
// ability to check free bet deposit amount.
func GetUserFreeBinaryOptionBets(tx *gorm.DB, userID int64) (*[]BenefitProgressBinaryOption, error) {
	if tx == nil {
		tx = db.DB
	}

	var benefitProgressIDs []int64
	err := tx.Model(&BenefitProgress{}).Where(
		"user_id = ? and polymorphic_benefit_progress_type = ?",
		userID, BenefitProgressBinaryOptionType).
		Pluck("polymorphic_benefit_progress_id", &benefitProgressIDs).Error
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	if len(benefitProgressIDs) == 0 {
		return nil, nil
	}

	var benefitProgressBOs []BenefitProgressBinaryOption
	err = tx.Find(&benefitProgressBOs, "id in ?", benefitProgressIDs).Error
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	return &benefitProgressBOs, nil
}
