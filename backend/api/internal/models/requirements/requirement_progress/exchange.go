package requirement_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const RequirementProgressExchangeType = "requirement_progress_exchange"

type RequirementProgressExchange struct {
	ID                     int64 `gorm:"primaryKey;autoIncrement"`
	CurrentBCoinsExchanged float64
}

// CreateRequirementProgressExchange creates RequirementProgressExchange and
// linked RequirementProgress. Requirement parameter should contain existing
// polymorphic requirement.
func CreateRequirementProgressExchange(tx *gorm.DB, requirement *requirements.Requirement, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	exchangeReqProgress := RequirementProgressExchange{}

	var requirementProgress RequirementProgress

	err := tx.Save(&exchangeReqProgress).Scan(&exchangeReqProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementProgress = RequirementProgress{
		UserID:                             userID,
		RequirementID:                      requirement.ID,
		PolymorphicRequirementProgressID:   exchangeReqProgress.ID,
		PolymorphicRequirementProgressType: RequirementProgressExchangeType,
		PolymorphicRequirementProgress:     exchangeReqProgress,
	}

	err = tx.Save(&requirementProgress).Scan(&requirementProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UpdateRequirementProgressExchangeIfRequired returns true if
// requirement completed and removes RequirementProgressExchange with
// RequirementProgress.
func UpdateRequirementProgressExchangeIfRequired(
	tx *gorm.DB, userID int64, bcoinsExchanged float64) (bool, error) {
	if tx == nil {
		tx = db.DB
	}

	var requirementProgress RequirementProgress
	err := tx.Preload("Requirement").First(&requirementProgress,
		"user_id = ? and polymorphic_requirement_progress_type = ?",
		userID, RequirementProgressExchangeType).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil // requirement not set
	} else if err != nil {
		return false, logger.WrapError(err, "")
	}

	// if set, update progress
	if err = requirementProgress.PreloadPolymorphicRequirementProgress(tx); err != nil {
		return false, logger.WrapError(err, "")
	}
	if err = requirementProgress.Requirement.PreloadPolymorphicRequirement(tx); err != nil {
		return false, logger.WrapError(err, "")
	}

	requirementProgressExchange, ok := requirementProgress.
		PolymorphicRequirementProgress.(RequirementProgressExchange)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirementProgress to (RequirementProgressExchange)"), "")
	}
	requirementExchange, ok := requirementProgress.Requirement.
		PolymorphicRequirement.(requirements.RequirementExchange)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirement to (RequirementExchange)"), "")
	}

	requirementProgressExchange.CurrentBCoinsExchanged += bcoinsExchanged
	if requirementProgressExchange.CurrentBCoinsExchanged >=
		requirementExchange.BCoinsAmount {
		// requirement completed
		if err = tx.Delete(&requirementProgressExchange).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		if err = tx.Delete(&requirementProgress).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		return true, nil
	} else if err = tx.Save(&requirementProgressExchange).Error; err != nil {
		return false, logger.WrapError(err, "")
	}

	return false, nil
}
