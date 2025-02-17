package requirements

import (
	"BlessedApi/cmd/db"
	"BlessedApi/pkg/logger"
	"fmt"

	"gorm.io/gorm"
)

type Requirement struct {
	ID                         int64       `gorm:"primaryKey;autoIncrement"`
	PolymorphicRequirementID   int64       `gorm:"index"`
	PolymorphicRequirementType string      `gorm:"index"`
	PolymorphicRequirement     interface{} `gorm:"-"`
}

// Loads requirement from db by its type and id. Errors not for export
func (req *Requirement) PreloadPolymorphicRequirement(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	var err error
	switch req.PolymorphicRequirementType {
	case RequirementMiniGameType:
		var requirementMiniGame RequirementMiniGame
		err = tx.First(&requirementMiniGame, req.PolymorphicRequirementID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		req.PolymorphicRequirement = requirementMiniGame
	case RequirementClickerType:
		var requirementClicker RequirementClicker
		err = tx.First(&requirementClicker, req.PolymorphicRequirementID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		req.PolymorphicRequirement = requirementClicker
	case RequirementExchangeType:
		var requirementExchange RequirementExchange
		err = tx.First(&requirementExchange, req.PolymorphicRequirementID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		req.PolymorphicRequirement = requirementExchange
	case RequirementReplenishmentType:
		var requirementReplenishment RequirementReplenishment
		err = tx.First(&requirementReplenishment, req.PolymorphicRequirementID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		req.PolymorphicRequirement = requirementReplenishment
	case RequirementTurnoverType:
		var requirementTurnover RequirementTurnover
		err = tx.First(&requirementTurnover, req.PolymorphicRequirementID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		req.PolymorphicRequirement = requirementTurnover
	case RequirementBinaryOptionType:
		var requirementBinaryOption RequirementBinaryOption
		err = tx.First(&requirementBinaryOption, req.PolymorphicRequirementID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		req.PolymorphicRequirement = requirementBinaryOption
	default:
		return logger.WrapError(err, fmt.Sprintf("no such PolymorphicRequirementType: %s", req.PolymorphicRequirementType))
	}

	return nil
}
