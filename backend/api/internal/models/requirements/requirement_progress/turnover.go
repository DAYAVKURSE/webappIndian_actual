package requirement_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/requirements"
	"BlessedApi/pkg/logger"
	"errors"
	"time"

	"gorm.io/gorm"
)

const RequirementProgressTurnoverType = "requirement_progress_turnover"

type RequirementProgressTurnover struct {
	ID                   int64 `gorm:"primaryKey;autoIncrement"`
	CurrentRupeeTurnover float64
	StartedAt            time.Time
}

// CreateRequirementProgressTurnover creates RequirementProgressTurnover and
// linked RequirementProgress. Requirement parameter should contain existing
// polymorphic requirement.
func CreateRequirementProgressTurnover(tx *gorm.DB, requirement *requirements.Requirement, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	turnoverReqProgress := RequirementProgressTurnover{
		StartedAt: time.Now()}

	var requirementProgress RequirementProgress

	err := tx.Save(&turnoverReqProgress).Scan(&turnoverReqProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	requirementProgress = RequirementProgress{
		UserID:                             userID,
		RequirementID:                      requirement.ID,
		PolymorphicRequirementProgressID:   turnoverReqProgress.ID,
		PolymorphicRequirementProgressType: RequirementProgressTurnoverType,
		PolymorphicRequirementProgress:     turnoverReqProgress,
	}

	err = tx.Save(&requirementProgress).Scan(&requirementProgress).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UpdateRequirementProgressTurnoverIfRequired returns true if
// requirement completed and removes RequirementProgressTurnover with
// RequirementProgress.
func UpdateRequirementProgressTurnoverIfRequired(tx *gorm.DB, userID int64, rupeeSpent float64) (bool, error) {
	if tx == nil {
		tx = db.DB
	}

	// check if requirement set
	var requirementProgress RequirementProgress
	err := tx.Preload("Requirement").First(&requirementProgress,
		"user_id = ? and polymorphic_requirement_progress_type = ?", userID, RequirementProgressTurnoverType).Error
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

	requirementProgressTurnover, ok := requirementProgress.
		PolymorphicRequirementProgress.(RequirementProgressTurnover)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirementProgress to (RequirementProgressTurnover)"), "")
	}
	requirementTurnover, ok := requirementProgress.Requirement.
		PolymorphicRequirement.(requirements.RequirementTurnover)
	if !ok {
		return false, logger.WrapError(errors.New(
			"unable to convert PolymorphicRequirement to (RequirementTurnover)"), "")
	}

	startedAt := requirementProgressTurnover.StartedAt
	timeDuration := time.Duration(requirementTurnover.TimeDuration) * time.Second

	// Progress expired
	if timeDuration != 0 && startedAt.Add(timeDuration).Before(time.Now()) {
		requirementProgressTurnover.CurrentRupeeTurnover = rupeeSpent
		requirementProgressTurnover.StartedAt = time.Now()
	} else {
		requirementProgressTurnover.CurrentRupeeTurnover += rupeeSpent
	}

	if requirementProgressTurnover.CurrentRupeeTurnover >= requirementTurnover.AmountRupee {
		// requirement completed
		if err = tx.Delete(&requirementProgressTurnover).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		if err = tx.Delete(&requirementProgress).Error; err != nil {
			return false, logger.WrapError(err, "")
		}

		return true, nil
	} else if err = tx.Save(&requirementProgressTurnover).Error; err != nil {
		return false, logger.WrapError(err, "")
	}

	return false, nil
}
