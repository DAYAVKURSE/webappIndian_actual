package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models/requirements/requirement_progress"
	"BlessedApi/pkg/logger"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserRequirementsProgress(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var reqProgs []requirement_progress.RequirementProgress
	errNotFound := errors.New("requirement progresses not found")

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		err = db.DB.Preload("Requirement").Find(&reqProgs, "user_id = ?", userID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}

		// Should happen only when user on 60lvl
		if len(reqProgs) == 0 {
			return errNotFound
		}

		for i := range reqProgs {
			err = reqProgs[i].Requirement.PreloadPolymorphicRequirement(nil)
			if err != nil {
				return logger.WrapError(err, "")
			}
			err = reqProgs[i].PreloadPolymorphicRequirementProgress(nil)
			if err != nil {
				return logger.WrapError(err, "")
			}
		}

		return nil
	})
	if err != nil && errors.Is(err, errNotFound) {
		c.String(404, "[]")
		return
	} else if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.JSON(200, reqProgs)
}
