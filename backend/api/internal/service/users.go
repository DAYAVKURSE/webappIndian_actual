package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/travepass"
	"BlessedApi/pkg/logger"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type signUpInput struct {
	Nickname string `validate:"required,min=3,max=32"`
	AvatarID int    `validate:"required,min=1,max=100"`
}

func (i *signUpInput) Validate() error {
	validate = validator.New()
	return validate.Struct(i)
}

func SignUp(c *gin.Context) {
	var input signUpInput
	var err error

	if err = c.Bind(&input); err != nil {
		c.JSON(400, gin.H{"error": "Unable to unmarshal body"})
		return
	}

	var user models.User
	if user.ID, err = middleware.GetUserIDFromGinContext(c); err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	user.AvatarID = input.AvatarID
	user.Nickname = input.Nickname

	exists, err := models.CheckIfUserExistsByID(user.ID)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}
	if exists {
		c.JSON(409, gin.H{"error": "User with this ID already exists"})
		return
	}

	exists, err = models.CheckIfUserExistsByNickname(user.Nickname)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}
	if exists {
		c.JSON(409, gin.H{"error": "User with this nickname already exists"})
		return
	}

	user.TravePassLevelID = 0

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(&user).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if referralID, ok := c.GetQuery("referral"); ok {
			if referrerID, err := strconv.ParseInt(referralID, 10, 64); err == nil {
				if err = tx.Create(&models.UserReferral{
					ReferrerID:       referrerID,
					ReferredID:       user.ID,
					ReferredNickname: input.Nickname,
				}).Error; err != nil {
					return logger.WrapError(err, "")
				}
			}
		}

		if err = travepass.CreateUserRequirementProgresses(
			tx, user.ID, user.TravePassLevelID+1); err != nil {
			return logger.WrapError(err, "")
		}
		return nil
	})
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.Status(200)
}

func GetUser(c *gin.Context) {
	var user models.User
	var err error

	user.ID, err = middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	err = db.DB.First(&user, user.ID).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	err = user.ResetClicksIfNecessaryAndSave()
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.JSON(200, user)
}

func GetUserReferrals(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	var userReferrals []models.UserReferral
	// Query the database for user referrals
	err = db.DB.Preload("ReferredFirstDeposit").
		Find(&userReferrals, "referrer_id = ?", userID).Error
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	if len(userReferrals) == 0 {
		c.JSON(404, userReferrals)
		return
	}

	c.JSON(200, userReferrals)
}

func Auth(c *gin.Context) {
	userId, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	exists, err := models.CheckIfUserExistsByID(userId)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	if exists {
		c.Status(200)
		return
	} else {
		c.Status(401)
		return
	}
}
