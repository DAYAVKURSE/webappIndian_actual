package benefit_progress

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/pkg/logger"
	"errors"

	"gorm.io/gorm"
)

const BenefitProgressMiniGameType = "benefit_mini_game_progress"

type BenefitProgressMiniGame struct {
	ID                  int64 `gorm:"primaryKey;autoIncrement"`
	GameID              int64 // gorm:"index"
	FreeBetsAmount      int
	FreeBetDepositRupee float64
}

// CreateBenefitProgressMiniGame creates BenefitProgressMiniGame and
// linked BenefitProgress. Benefit parameter should contain existing
// polymorphic benefit.
func CreateBenefitProgressMiniGame(tx *gorm.DB, benefit *benefits.Benefit, userID int64) error {
	if tx == nil {
		tx = db.DB
	}

	benefitMiniGame, ok := benefit.PolymorphicBenefit.(benefits.BenefitMiniGame)
	if !ok {
		return logger.WrapError(errors.New(
			"unable to cast benefit.PolymorphicBenefit to BenefitMiniGame"), "")
	}

	BenefitProgressMiniGame := BenefitProgressMiniGame{
		GameID:              benefitMiniGame.GameID,
		FreeBetsAmount:      benefitMiniGame.FreeBetsAmount,
		FreeBetDepositRupee: benefitMiniGame.FreeBetDepositRupee,
	}

	err := tx.Save(&BenefitProgressMiniGame).
		Scan(&BenefitProgressMiniGame).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	err = tx.Save(&BenefitProgress{
		UserID:                         userID,
		BenefitID:                      benefit.ID,
		PolymorphicBenefitProgressID:   BenefitProgressMiniGame.ID,
		PolymorphicBenefitProgressType: BenefitProgressMiniGameType}).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

// UseFreeMiniGameBetIfAvailable checks if there are available benefits
// on mini game by game id for user, and returns free bet deposit if bet is available
// with function to update free bets count. If last free bet used,
// BenefitProgressMiniGame with BenefitProgress will be deleted.
func UseFreeMiniGameBetIfAvailable(tx *gorm.DB, userID, gameID int64) (
	float64, func(tx *gorm.DB) error, error) {
	if tx == nil {
		tx = db.DB
	}

	var benefitProgresses []BenefitProgress
	err := tx.Find(&benefitProgresses,
		"user_id = ? and polymorphic_benefit_progress_type = ?",
		userID, BenefitProgressMiniGameType).Error
	if err != nil {
		return 0, nil, logger.WrapError(err, "")
	}

	if len(benefitProgresses) == 0 {
		return 0, nil, nil
	}

	for i := range benefitProgresses {
		if err = benefitProgresses[i].PreloadPolymorphicBenefitProgress(tx); err != nil {
			return 0, nil, logger.WrapError(err, "")
		}
		benefitProgressMiniGame, ok := benefitProgresses[i].
			PolymorphicBenefitProgress.(BenefitProgressMiniGame)
		if !ok {
			return 0, nil, logger.WrapError(errors.New(
				"unable to cast PolymorphicBenefitProgress to BenefitProgressMiniGame"), "")
		}

		// get first benefit progress with given gameID
		if benefitProgressMiniGame.GameID == gameID {

			return benefitProgressMiniGame.FreeBetDepositRupee, func(tx *gorm.DB) error {
				benefitProgressMiniGame.FreeBetsAmount--

				if benefitProgressMiniGame.FreeBetsAmount == 0 {
					if err = tx.Delete(&benefitProgressMiniGame).Error; err != nil {
						return logger.WrapError(err, "")
					}
					if err = tx.Delete(&benefitProgresses[i]).Error; err != nil {
						return logger.WrapError(err, "")
					}
				} else if err = tx.Save(&benefitProgressMiniGame).Error; err != nil {
					return logger.WrapError(err, "")
				}
				return nil
			}, nil
		}
	}

	return 0, nil, nil
}

func GetUserFreeMiniGameBets(tx *gorm.DB, userID, gameID int64) (*[]BenefitProgressMiniGame, error) {
	if tx == nil {
		tx = db.DB
	}

	var benefitProgressIDs []int64
	err := tx.Model(&BenefitProgress{}).Where(
		"user_id = ? and polymorphic_benefit_progress_type = ?",
		userID, BenefitProgressMiniGameType).
		Pluck("polymorphic_benefit_progress_id", &benefitProgressIDs).Error
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	if len(benefitProgressIDs) == 0 {
		return nil, nil
	}

	var benefitProgressMGs []BenefitProgressMiniGame
	err = tx.Find(&benefitProgressMGs,
		"id in ? and game_id = ?",
		benefitProgressIDs, gameID).Error
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	return &benefitProgressMGs, nil
}
