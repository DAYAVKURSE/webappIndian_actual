package requirement_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const RequirementProgressMiniGameType = "requirement_progress_mini_game"

type RequirementProgressMiniGame struct {
	ID         int64 `gorm:"primaryKey;autoIncrement"`
	BetsAmount int
	WinsAmount int
}

// CreateRequirementProgressMiniGame creates RequirementProgressMiniGame and
// linked RequirementProgress. Requirement parameter should contain existing
// polymorphic requirement.
func CreateRequirementProgressMiniGame(tx *gorm.DB, requirement *requirements.Requirement, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	gamesReqProgress := RequirementProgressMiniGame{}

	var requirementProgress RequirementProgress
	err := tx.Save(&gamesReqProgress).Scan(&gamesReqProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementProgress = RequirementProgress{
		UserID:                             userID,
		RequirementID:                      requirement.ID,
		PolymorphicRequirementProgressID:   gamesReqProgress.ID,
		PolymorphicRequirementProgressType: RequirementProgressMiniGameType,
		PolymorphicRequirementProgress:     gamesReqProgress,
	}

	err = tx.Save(&requirementProgress).Scan(&requirementProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UpdateRequirementProgressMiniGameIfRequired returns true if
// requirement completed and removes RequirementProgressMiniGame with
// RequirementProgress.
func UpdateRequirementProgressMiniGameIfRequired(
	tx *gorm.DB, userID, gameID int64, betAmount, betPayout float64) (bool, error) {
	if tx == nil {
		tx = db.DB
	}

	// check if requirement set
	var requirementProgress RequirementProgress
	err := tx.Preload("Requirement").First(&requirementProgress,
		"user_id = ? and polymorphic_requirement_progress_type = ?", userID, RequirementProgressMiniGameType).Error
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

	requirementProgressMiniGame, ok := requirementProgress.
		PolymorphicRequirementProgress.(RequirementProgressMiniGame)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirementProgress to (RequirementProgressReplenishment)"), "")
	}
	requirementMiniGame, ok := requirementProgress.Requirement.
		PolymorphicRequirement.(requirements.RequirementMiniGame)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirement to (RequirementReplenishment)"), "")
	}

	if requirementMiniGame.GameID != 0 && requirementMiniGame.GameID != gameID {
		return false, nil
	}

	if betAmount >= requirementMiniGame.MinBetRupee {
		requirementProgressMiniGame.BetsAmount++
		if betPayout > 0 {
			requirementProgressMiniGame.WinsAmount++
		}
	} else {
		return false, nil
	}

	if requirementProgressMiniGame.BetsAmount >= requirementMiniGame.BetsAmount &&
		requirementProgressMiniGame.WinsAmount >= requirementMiniGame.WinsAmount {
		// requirement completed
		if err = tx.Delete(&requirementProgressMiniGame).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		if err = tx.Delete(&requirementProgress).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		return true, nil
	} else if err = tx.Save(&requirementProgressMiniGame).Error; err != nil {
		return false, logger.WrapError(err, "")
	}

	return false, nil
}
