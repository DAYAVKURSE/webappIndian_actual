package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/logger"

	"github.com/gin-gonic/gin"
)

func GetAllTravePassLevelsWithRequirementsAndBenefits(c *gin.Context) {
	levels, err := travepass.GetAllTravePassLevelsWithRequirementsAndBenefits(nil)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	if len(*levels) == 0 {
		logger.Error("Trave pass levels absent")
		c.String(404, "[]")
		return
	}

	c.JSON(200, *levels)
}

func GetNextLevelRequirements(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var travePassNextLevelID int64
	err = db.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Pluck("trave_pass_level_id", &travePassNextLevelID).Error
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Get user next level
	travePassNextLevelID++

	var reqs []travepass.TravePassLevelRequirement
	err = db.DB.Preload("Requirement").Find(&reqs, "trave_pass_level_id = ?", travePassNextLevelID).Error
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Can only happen on 60lvl
	if len(reqs) == 0 {
		c.String(404, "[]")
		return
	}

	for i := range reqs {
		err = reqs[i].Requirement.PreloadPolymorphicRequirement(nil)
		if err != nil {
			logger.Error("UserID: %d; level requirements empty", userID)
			c.Status(500)
			return
		}
	}

	c.JSON(200, reqs)
}
