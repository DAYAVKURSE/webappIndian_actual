package requirement_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"errors"
	"time"

	"gorm.io/gorm"
)

const RequirementProgressClickerType = "requirement_progress_clicker"

type RequirementProgressClicker struct {
	ID            int64 `gorm:"primaryKey;autoIncrement"`
	CurrentClicks int
	StartedAt     time.Time
}

// CreateRequirementProgressClicker creates RequirementProgressClicker and
// linked RequirementProgress. Requirement parameter should contain existing
// polymorphic requirement.
func CreateRequirementProgressClicker(tx *gorm.DB, requirement *requirements.Requirement, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	clickerReqProgress := RequirementProgressClicker{
		StartedAt: time.Now()}

	var requirementProgress RequirementProgress

	err := tx.Save(&clickerReqProgress).Scan(&clickerReqProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementProgress = RequirementProgress{
		UserID:                             userID,
		RequirementID:                      requirement.ID,
		PolymorphicRequirementProgressID:   clickerReqProgress.ID,
		PolymorphicRequirementProgressType: RequirementProgressClickerType,
		PolymorphicRequirementProgress:     clickerReqProgress,
	}

	err = tx.Save(&requirementProgress).Scan(&requirementProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UpdateRequirementProgressClickerIfRequired returns true if
// requirement completed and removes RequirementProgressClicker with
// RequirementProgress.
func UpdateRequirementProgressClickerIfRequired(tx *gorm.DB, userID int64, clicksCount int) (bool, error) {
	// check if requirement set
	if tx == nil {
		tx = db.DB
	}

	var requirementProgress RequirementProgress
	err := tx.Preload("Requirement").First(&requirementProgress,
		"user_id = ? and polymorphic_requirement_progress_type = ?", userID, RequirementProgressClickerType).Error
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

	requirementProgressClicker, ok := requirementProgress.
		PolymorphicRequirementProgress.(RequirementProgressClicker)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirementProgress to (RequirementProgressClicker)"), "")
	}
	requirementClicker, ok := requirementProgress.Requirement.
		PolymorphicRequirement.(requirements.RequirementClicker)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirement to (RequirementClicker)"), "")
	}

	// if requirement is hit daily clicks limit
	if requirementClicker.HitLimit {
		var userDailyClicks int
		err := tx.Model(&models.User{}).Where("id = ?", userID).Pluck("daily_clicks", &userDailyClicks).Error
		if err != nil {
			return false, logger.WrapError(err, "")
		}

		if userDailyClicks == models.DailyClicksLimit {
			// requirement completed
			if err = tx.Delete(&requirementProgressClicker).Error; err != nil {
				return false, logger.WrapError(err, "")
			}

			if err = tx.Delete(&requirementProgress).Error; err != nil {
				return false, logger.WrapError(err, "")
			}

			return true, nil
		}
	}

	startedAt := requirementProgressClicker.StartedAt
	timeDuration := time.Duration(requirementClicker.TimeDuration) * time.Second

	// User progress expired
	if timeDuration != 0 && startedAt.Add(timeDuration).Before(time.Now()) {
		requirementProgressClicker.CurrentClicks = clicksCount
		requirementProgressClicker.StartedAt = time.Now()
	} else {
		requirementProgressClicker.CurrentClicks += clicksCount
	}

	if requirementProgressClicker.CurrentClicks >= requirementClicker.ClicksAmount {
		// requirement completed
		if err = tx.Delete(&requirementProgressClicker).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		if err = tx.Delete(&requirementProgress).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		return true, nil
	} else if err = tx.Save(&requirementProgressClicker).Error; err != nil {
		return false, logger.WrapError(err, "")
	}

	return false, nil
}
