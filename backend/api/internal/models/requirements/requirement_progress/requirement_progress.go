package requirement_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"fmt"

	"gorm.io/gorm"
)

type RequirementProgress struct {
	ID                                 int64                    `gorm:"primaryKey;autoIncrement"`
	UserID                             int64                    `gorm:"index"`
	RequirementID                      int64                    `gorm:"index"`
	Requirement                        requirements.Requirement `gorm:"foreignKey:RequirementID;constraint:OnDelete:CASCADE;"`
	PolymorphicRequirementProgressID   int64                    `gorm:"index"`
	PolymorphicRequirementProgressType string                   `gorm:"index"`
	PolymorphicRequirementProgress     interface{}              `gorm:"-"`
}

// PreloadPolymorphicRequirementProgress preloads RequirementProgress
// polymorphic relation PolymorphicRequirementProgress by its type and id.
func (rp *RequirementProgress) PreloadPolymorphicRequirementProgress(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	var err error
	switch rp.PolymorphicRequirementProgressType {
	case RequirementProgressClickerType:
		var requirementProgressClicker RequirementProgressClicker
		err = tx.First(
			&requirementProgressClicker, rp.PolymorphicRequirementProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		rp.PolymorphicRequirementProgress = requirementProgressClicker
	case RequirementProgressExchangeType:
		var requirementProgressExchange RequirementProgressExchange
		err = tx.First(
			&requirementProgressExchange, rp.PolymorphicRequirementProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		rp.PolymorphicRequirementProgress = requirementProgressExchange
	case RequirementProgressMiniGameType:
		var requirementProgressMiniGame RequirementProgressMiniGame
		err = tx.First(
			&requirementProgressMiniGame, rp.PolymorphicRequirementProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		rp.PolymorphicRequirementProgress = requirementProgressMiniGame
	case RequirementProgressReplenishmentType:
		var requirementProgressReplenishment RequirementProgressReplenishment
		err = tx.First(
			&requirementProgressReplenishment, rp.PolymorphicRequirementProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		rp.PolymorphicRequirementProgress = requirementProgressReplenishment
	case RequirementProgressTurnoverType:
		var requirementProgressTurnover RequirementProgressTurnover
		err = tx.First(
			&requirementProgressTurnover, rp.PolymorphicRequirementProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		rp.PolymorphicRequirementProgress = requirementProgressTurnover
	case RequirementProgressBinaryOptionType:
		var requirementProgressBinaryOption RequirementProgressBinaryOption
		err = tx.First(
			&requirementProgressBinaryOption, rp.PolymorphicRequirementProgressID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}
		rp.PolymorphicRequirementProgress = requirementProgressBinaryOption
	default:
		return logger.WrapError(err, fmt.Sprintf(
			"no such PolymorphicRequirementProgressType: %s",
			rp.PolymorphicRequirementProgressType))
	}

	return nil
}

// CreatePolymorphicRequirementProgress creates polymorphic requirement
// progress and linked requirement progress. Requirement parameter should
// contain existing polymorphic requirement.
func CreatePolymorphicRequirementProgress(
	tx *gorm.DB, requirement *requirements.Requirement, userID int64) (err error) {
	switch requirement.PolymorphicRequirementType {
	case requirements.RequirementClickerType:
		if err = CreateRequirementProgressClicker(
			tx, requirement, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case requirements.RequirementExchangeType:
		if err = CreateRequirementProgressExchange(
			tx, requirement, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case requirements.RequirementMiniGameType:
		if err = CreateRequirementProgressMiniGame(
			tx, requirement, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case requirements.RequirementReplenishmentType:
		if err = CreateRequirementProgressReplenishment(
			tx, requirement, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case requirements.RequirementTurnoverType:
		if err = CreateRequirementProgressTurnover(
			tx, requirement, userID); err != nil {
			return logger.WrapError(err, "")
		}
	case requirements.RequirementBinaryOptionType:
		if err = CreateRequirementProgressBinaryOption(
			tx, requirement, userID); err != nil {
			return logger.WrapError(err, "")
		}
	default:
		return logger.WrapError(err, fmt.Sprintf(
			"no such polymorphicRequirementType: %s",
			requirement.PolymorphicRequirementType))
	}
	return nil
}
