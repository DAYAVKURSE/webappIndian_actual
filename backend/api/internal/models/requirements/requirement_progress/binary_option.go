package requirement_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const RequirementProgressBinaryOptionType = "requirement_progress_binary_option"

type RequirementProgressBinaryOption struct {
	ID                 int64 `gorm:"primaryKey;autoIncrement"`
	BetsAmount         int
	WinsAmount         int
	TotalWinningsRupee float64
}

// CreateRequirementProgressBinaryOption creates RequirementProgressBinaryOption and
// linked RequirementProgress. Requirement parameter should contain existing
// polymorphic requirement.
func CreateRequirementProgressBinaryOption(tx *gorm.DB, requirement *requirements.Requirement, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	binaryOptionReqProgress := RequirementProgressBinaryOption{}

	var requirementProgress RequirementProgress

	err := tx.Save(&binaryOptionReqProgress).Scan(&binaryOptionReqProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementProgress = RequirementProgress{
		UserID:                             userID,
		RequirementID:                      requirement.ID,
		PolymorphicRequirementProgressID:   binaryOptionReqProgress.ID,
		PolymorphicRequirementProgressType: RequirementProgressBinaryOptionType,
		PolymorphicRequirementProgress:     binaryOptionReqProgress,
	}

	err = tx.Save(&requirementProgress).Scan(&requirementProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UpdateRequirementProgressBinaryOptionIfRequired returns true if
// requirement completed and removes RequirementProgressBinaryOption with
// RequirementProgress.
func UpdateRequirementProgressBinaryOptionIfRequired(
	tx *gorm.DB, userID int64, betAmount, betPayout float64) (bool, error) {
	if tx == nil {
		tx = db.DB
	}

	// check if requirement set
	var requirementProgress RequirementProgress
	err := tx.Preload("Requirement").First(&requirementProgress,
		"user_id = ? and polymorphic_requirement_progress_type = ?", userID, RequirementProgressBinaryOptionType).Error
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

	requirementProgressBinaryOption, ok := requirementProgress.
		PolymorphicRequirementProgress.(RequirementProgressBinaryOption)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirementProgress to (RequirementProgressBinaryOption)"), "")
	}
	requirementBinaryOption, ok := requirementProgress.Requirement.
		PolymorphicRequirement.(requirements.RequirementBinaryOption)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirement to (RequirementBinaryOption)"), "")
	}

	if betAmount <
		requirementBinaryOption.MinBetRupee {
		return false, nil
	}

	requirementProgressBinaryOption.BetsAmount++
	if betPayout > 0 {
		requirementProgressBinaryOption.WinsAmount++
		requirementProgressBinaryOption.TotalWinningsRupee += betPayout
	}

	if requirementProgressBinaryOption.BetsAmount >=
		requirementBinaryOption.BetsAmount &&
		requirementProgressBinaryOption.WinsAmount >=
			requirementBinaryOption.WinsAmount &&
		requirementProgressBinaryOption.TotalWinningsRupee >=
			requirementBinaryOption.TotalWinningsRupee {
		// requirement completed
		if err = tx.Delete(&requirementProgressBinaryOption).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		if err = tx.Delete(&requirementProgress).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		return true, nil
	} else if err = tx.Save(&requirementProgressBinaryOption).Error; err != nil {
		return false, logger.WrapError(err, "")
	}

	return false, nil
}
