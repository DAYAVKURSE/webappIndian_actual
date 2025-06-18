package models

import (
	"BlessedApi/cmd/db"
	"BlessedApi/pkg/logger"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

const DailyClicksLimit = 10000
const MinBiPerClick = 0

type User struct {
	ID               int64  `gorm:"primaryKey,autoIncrement"`
	Nickname         string `gorm:"unique"`
	AvatarID         int
	BalanceBi        float64
	BalanceRupee     float64
	TurnoverRupee    float64
	DailyClicks      int
	TravePassLevelID int64 `gorm:"index;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	LastClicksReset  time.Time
	CreatedAt        time.Time
	Password         string
}

func (u *User) Validate() error {
	return validate.Struct(u)
}

func (u *User) ResetClicksIfNecessaryAndSave() error {
	now := time.Now()
	// Calculate the start of today (00:00)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if u.LastClicksReset.Before(startOfToday) {
		u.LastClicksReset = now
		u.DailyClicks = 0
	}

	err := db.DB.Save(&u).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

func CheckIfUserExistsByID(userID int64) (bool, error) {
	var exists bool
	err := db.DB.Model(&User{}).
		Select("count(*) > 0").
		Where("id = ?", userID).
		Scan(&exists).Error
	if err != nil {
		return true, logger.WrapError(err, "")
	}

	return exists, nil
}

func GetUserWithPassword(nickname string) (*User, error) {
	var user User

	err := db.DB.
		Where("nickname = ?", nickname).
		First(&user).Error
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	return &user, nil
}

func CheckIfUserExistsByNickname(nn string) (bool, error) {
	var exists bool

	err := db.DB.Model(&User{}).
		Select("count(*) > 0").
		Where("nickname = ?", nn).
		Scan(&exists).Error
	if err != nil {
		return true, logger.WrapError(err, "")
	}

	return exists, nil
}

func CountUserBiPerClick(userId int64) (float64, error) {
	var lastDeps []Deposit
	err := db.DB.Where("user_id = ? AND bonus_expires_in > ?", userId, time.Now()).Find(&lastDeps).Error
	if err != nil {
		return MinBiPerClick, logger.WrapError(err, "")
	}

	if len(lastDeps) == 0 {
		return MinBiPerClick, nil
	}

	var biPerClick float64
	for _, v := range lastDeps {
		biPerClick += (v.AmountRupee * 0.1) * float64(BiRupeeCourse) / float64(DailyClicksLimit)
	}

	return biPerClick, nil
}
