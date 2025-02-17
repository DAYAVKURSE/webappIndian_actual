package requirement_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const RequirementProgressReplenishmentType = "requirement_progress_replenishment"

type RequirementProgressReplenishment struct {
	ID                        int64 `gorm:"primaryKey;autoIncrement"`
	CurrentReplenishmentRupee float64
}

// CreateRequirementProgressReplenishment creates RequirementProgressReplenishment and
// linked RequirementProgress. Requirement parameter should contain existing
// polymorphic requirement.
func CreateRequirementProgressReplenishment(tx *gorm.DB, requirement *requirements.Requirement, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	replenishmentReqProgress := RequirementProgressReplenishment{}

	var requirementProgress RequirementProgress
	err := tx.Save(&replenishmentReqProgress).Scan(&replenishmentReqProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementProgress = RequirementProgress{
		UserID:                             userID,
		RequirementID:                      requirement.ID,
		PolymorphicRequirementProgressID:   replenishmentReqProgress.ID,
		PolymorphicRequirementProgressType: RequirementProgressReplenishmentType,
		PolymorphicRequirementProgress:     replenishmentReqProgress,
	}

	err = tx.Save(&requirementProgress).Scan(&requirementProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UpdateRequirementProgressReplenishmentIfRequired returns true if
// requirement completed and removes RequirementProgressReplenishment with
// RequirementProgress.
func UpdateRequirementProgressReplenishmentIfRequired(
	tx *gorm.DB, userID int64, replenishmentRupeeValue float64) (bool, error) {
	if tx == nil {
		tx = db.DB
	}

	// check if requirement set
	var requirementProgress RequirementProgress
	err := tx.Preload("Requirement").First(&requirementProgress,
		"user_id = ? and polymorphic_requirement_progress_type = ?",
		userID, RequirementProgressReplenishmentType).Error
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

	requirementProgressReplenishment, ok := requirementProgress.
		PolymorphicRequirementProgress.(RequirementProgressReplenishment)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirementProgress to (RequirementProgressReplenishment)"), "")
	}
	requirementReplenishment, ok := requirementProgress.Requirement.
		PolymorphicRequirement.(requirements.RequirementReplenishment)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirement to (RequirementReplenishment)"), "")
	}

	requirementProgressReplenishment.CurrentReplenishmentRupee +=
		replenishmentRupeeValue
	if requirementProgressReplenishment.CurrentReplenishmentRupee >=
		requirementReplenishment.AmountRupee {
		// requirement completed
		if err = tx.Delete(&requirementProgressReplenishment).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		if err = tx.Delete(&requirementProgress).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		return true, nil
	} else if err = tx.Save(&requirementProgressReplenishment).Error; err != nil {
		return false, logger.WrapError(err, "")
	}

	return false, nil
}
