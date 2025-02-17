package travepass

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/pkg/logger"

	"gorm.io/gorm"
)

type TravePassLevelRequirement struct {
	ID               int64                    `gorm:"primaryKey;autoIncrement"`
	TravePassLevelID int64                    `gorm:"index"`
	RequirementID    int64                    `gorm:"index"`
	Requirement      requirements.Requirement `gorm:"foreignKey:RequirementID;constraint:OnDelete:SET NULL;"`
}

// CreateUserRequirementProgresses creates RequirementProgresses with trave pass next level id.
// Should be used on trave pass level up.
func CreateUserRequirementProgresses(tx *gorm.DB, userID, nextLevelID int64) error {
	if tx == nil {
		tx = db.DB
	}

	var levelRequirements []TravePassLevelRequirement

	err := tx.Preload("Requirement").Find(&levelRequirements, "trave_pass_level_id = ?", nextLevelID).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	for reqIdx := range levelRequirements {
		if err = levelRequirements[reqIdx].Requirement.PreloadPolymorphicRequirement(tx); err != nil {
			return logger.WrapError(err, "")
		}
		if err = requirement_progress.CreatePolymorphicRequirementProgress(
			tx, &levelRequirements[reqIdx].Requirement, userID); err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}
