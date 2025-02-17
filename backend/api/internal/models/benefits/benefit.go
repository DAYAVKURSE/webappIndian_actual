package benefits

import (
	"BlessedApi/cmd/db"
	"BlessedApi/pkg/logger"

	"gorm.io/gorm"
)

type Benefit struct {
	ID                     int64       `gorm:"primaryKey;autoIncrement"`
	PolymorphicBenefitID   int64       `gorm:"index"`
	PolymorphicBenefitType string      `gorm:"index"`
	PolymorphicBenefit     interface{} `gorm:"-"`
}

// PreloadPolymorphicBenefit preloads Benefit
// polymorphic relation PolymorphicBenefit by its type and id.
func (ben *Benefit) PreloadPolymorphicBenefit(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	var err error
	switch ben.PolymorphicBenefitType {
	case BenefitClickerType:
		var clickerBen BenefitClicker
		err = tx.First(&clickerBen, ben.PolymorphicBenefitID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		ben.PolymorphicBenefit = clickerBen
	case BenefitCreditType:
		var creditBen BenefitCredit
		err = tx.First(&creditBen, ben.PolymorphicBenefitID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		ben.PolymorphicBenefit = creditBen
	case BenefitMiniGameType:
		var gameBen BenefitMiniGame
		err = tx.First(&gameBen, ben.PolymorphicBenefitID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		ben.PolymorphicBenefit = gameBen
	case BenefitItemType:
		var itemBen BenefitItem
		err = tx.First(&itemBen, ben.PolymorphicBenefitID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		ben.PolymorphicBenefit = itemBen
	case BenefitReplenishmentType:
		var replenishmentBen BenefitReplenishment
		err = tx.First(&replenishmentBen, ben.PolymorphicBenefitID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		ben.PolymorphicBenefit = replenishmentBen
	case BenefitBinaryOptionType:
		var binaryOptBen BenefitBinaryOption
		err = tx.First(&binaryOptBen, ben.PolymorphicBenefitID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		ben.PolymorphicBenefit = binaryOptBen
	case BenefitFortuneWheelType:
		var fortuneWheelBen BenefitFortuneWheel
		err = tx.First(&fortuneWheelBen, ben.PolymorphicBenefitID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		ben.PolymorphicBenefit = fortuneWheelBen
	}

	return nil
}
