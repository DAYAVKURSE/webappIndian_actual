package benefit_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/pkg/logger"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type BenefitProgress struct {
	ID                             int64            `gorm:"primaryKey;autoIncrement"`
	UserID                         int64            `gorm:"index"`
	BenefitID                      int64            `gorm:"index"`
	Benefit                        benefits.Benefit `gorm:"foreignKey:BenefitID;constraint:OnDelete:SET NULL;"`
	PolymorphicBenefitProgressID   int64            `gorm:"index"`
	PolymorphicBenefitProgressType string           `gorm:"index"`
	PolymorphicBenefitProgress     interface{}      `gorm:"-"`
}

// PreloadPolymorphicBenefitProgress preloads BenefitProgress
// polymorphic relation PolymorphicBenefitProgress by its type and id.
func (bp *BenefitProgress) PreloadPolymorphicBenefitProgress(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	var err error
	switch bp.PolymorphicBenefitProgressType {
	case BenefitProgressBinaryOptionType:
		var benBinOptProgress BenefitProgressBinaryOption
		err = tx.First(&benBinOptProgress, bp.PolymorphicBenefitProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		bp.PolymorphicBenefitProgress = benBinOptProgress
	case BenefitProgressClickerType:
		var benClickerProgress BenefitProgressClicker
		err = tx.First(&benClickerProgress, bp.PolymorphicBenefitProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		bp.PolymorphicBenefitProgress = benClickerProgress
	case BenefitProgressFortuneWheelType:
		var benFortWhProgress BenefitProgressFortuneWheel
		err = tx.First(&benFortWhProgress, bp.PolymorphicBenefitProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		bp.PolymorphicBenefitProgress = benFortWhProgress
	case BenefitProgressMiniGameType:
		var benMiniGameProgress BenefitProgressMiniGame
		err = tx.First(&benMiniGameProgress, bp.PolymorphicBenefitProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		bp.PolymorphicBenefitProgress = benMiniGameProgress
	case BenefitProgressReplenishmentType:
		var benReplProgress BenefitProgressReplenishment
		err = tx.First(&benReplProgress, bp.PolymorphicBenefitProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		bp.PolymorphicBenefitProgress = benReplProgress
	default:
		return logger.WrapError(err, fmt.Sprintf(
			"no such PolymorphicBenefitProgressType: %s",
			bp.PolymorphicBenefitProgressType))
	}

	return nil
}

// CreateOrApplyPolymorphicBenefitProgress creates polymorphic benefit
// progress and linked benefit progress. Benefit parameter should
// contain existing polymorphic benefit.
func CreateOrApplyPolymorphicBenefitProgress(tx *gorm.DB, benefit *benefits.Benefit, userID int64) (err error) {
	switch benefit.PolymorphicBenefitType {
	case benefits.BenefitBinaryOptionType:
		if err = CreateBenefitProgressBinaryOption(
			tx, benefit, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case benefits.BenefitClickerType:
		if err = CreateOrApplyBenefitProgressClicker(
			tx, benefit, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case benefits.BenefitCreditType:
		benefitCredit, ok := benefit.PolymorphicBenefit.(benefits.BenefitCredit)
		if !ok {
			return logger.WrapError(errors.New(
				"unable to cast benefit.PolymorphicBenefit to BenefitCredit"), "")
		}
		if err = benefitCredit.ApplyBenefit(tx, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case benefits.BenefitFortuneWheelType:
		if err = CreateBenefitProgressFortuneWheel(
			tx, benefit, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case benefits.BenefitItemType:
		benefitItem, ok := benefit.PolymorphicBenefit.(benefits.BenefitItem)
		if !ok {
			return logger.WrapError(errors.New(
				"unable to cast benefit.PolymorphicBenefit to BenefitItem"), "")
		}
		if err = benefitItem.ApplyBenefit(tx, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case benefits.BenefitMiniGameType:
		if err = CreateBenefitProgressMiniGame(
			tx, benefit, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case benefits.BenefitReplenishmentType:
		if err = CreateBenefitProgressReplenishment(
			tx, benefit, userID); err != nil {
			return logger.WrapError(err, "")
		}
	default:
		return logger.WrapError(err, fmt.Sprintf("no such PolymorphicBenefitType: %s", benefit.PolymorphicBenefitType))
	}

	return nil
}
