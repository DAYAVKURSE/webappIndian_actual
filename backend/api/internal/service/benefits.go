package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models/benefits/benefit_progress"
	"BlessedApi/pkg/logger"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserBenefitsProgress(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var benefitProgresses []benefit_progress.BenefitProgress
	errBenefitsNotFound := errors.New("benefits not found")

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Preload("Benefit").
			Find(&benefitProgresses, "user_id = ?", userID).Error
		if err != nil {
			return logger.WrapError(err, "")
		}

		if len(benefitProgresses) == 0 {
			return errBenefitsNotFound
		}

		for i := range benefitProgresses {
			err = benefitProgresses[i].Benefit.PreloadPolymorphicBenefit(tx)
			if err != nil {
				return logger.WrapError(err, "")
			}
			err = benefitProgresses[i].PreloadPolymorphicBenefitProgress(tx)
			if err != nil {
				return logger.WrapError(err, "")
			}
		}

		return nil
	})
	if err != nil && errors.Is(err, errBenefitsNotFound) {
		c.String(404, "[]")
	} else if err != nil {
		logger.Error("%v", err)
		return
	}

	c.JSON(200, benefitProgresses)
}
