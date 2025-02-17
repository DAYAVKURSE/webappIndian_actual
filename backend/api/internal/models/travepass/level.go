package travepass

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/pkg/logger"

	"gorm.io/gorm"
)

type TravePassLevel struct {
	ID           int64                       `gorm:"primaryKey;autoIncrement"`
	Requirements []TravePassLevelRequirement `gorm:"foreignKey:TravePassLevelID"`
	Benefits     []TravePassLevelBenefit     `gorm:"foreignKey:TravePassLevelID"`
}

// GetAllTravePassLevelsWithRequirementsAndBenefits loads all TravePassLevel with
// it's polymorphic relations.
func GetAllTravePassLevelsWithRequirementsAndBenefits(tx *gorm.DB) (*[]TravePassLevel, error) {
	if tx == nil {
		tx = db.DB
	}

	var levels []TravePassLevel
	err := tx.Preload("Benefits.Benefit").Preload("Benefits").
		Preload("Requirements.Requirement").Preload("Requirements").Find(&levels).Error
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	for levelIdx := range levels {
		for reqIdx := range levels[levelIdx].Requirements {
			if err = levels[levelIdx].Requirements[reqIdx].Requirement.PreloadPolymorphicRequirement(tx); err != nil {
				return nil, logger.WrapError(err, "")
			}
		}

		for benIdx := range levels[levelIdx].Benefits {
			if err = levels[levelIdx].Benefits[benIdx].Benefit.PreloadPolymorphicBenefit(tx); err != nil {
				return nil, logger.WrapError(err, "")
			}
		}
	}

	return &levels, nil
}

// CheckAndUpgradeTravePassLevel checks count of uncompleted trave pass level
// requirements by id of trave pass level. If all requirements completed(deleted)
// it applies PolymorphicBenefitProgresses and next level PolymorphicRequirementProgresses to user.
func CheckAndUpgradeTravePassLevel(tx *gorm.DB, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	var user models.User
	err := tx.First(&user, userID).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	// Get level requirements ids
	var nextLevelRequirementsIDs []int64
	err = tx.Model(&TravePassLevelRequirement{}).
		Where("trave_pass_level_id = ?", user.TravePassLevelID+1).
		Pluck("requirement_id", &nextLevelRequirementsIDs).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	// Get count of user LEVEL uncompleted requirements
	var requirementsProgressCount int
	err = tx.Model(&requirement_progress.RequirementProgress{}).
		Select("count(*)").
		Where("user_id = ? and requirement_id in ?", userID, nextLevelRequirementsIDs).
		Scan(&requirementsProgressCount).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	// if all requirements completed - give benefits, set new level and create new RequirementsProgresses
	if requirementsProgressCount == 0 {
		if user.TravePassLevelID == 60 {
			return nil
		}

		user.TravePassLevelID += 1

		if err = tx.Save(&user).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if err = CreateUserBenefitProgresses(
			tx, userID, user.TravePassLevelID); err != nil {
			return logger.WrapError(err, "")
		}

		if err = CreateUserRequirementProgresses(
			tx, userID, user.TravePassLevelID+1); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}
