package fortune_wheel

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models/benefits"
	"BlessedApi/pkg/logger"
	"errors"
	"math/rand"

	"gorm.io/gorm"
)

type FortuneWheelSector struct {
	ID          int64   `gorm:"primaryKey,autoIncrement"`
	Probability float64 `json:"-"`
	ColorHex    string
	BenefitID   int64            `gorm:"index"`
	Benefit     benefits.Benefit `gorm:"foreignKey:BenefitID;constraint:OnDelete:SET NULL;"`
}

func GetAllFortuneWheelSectors(tx *gorm.DB) ([]FortuneWheelSector, error) {
	if tx == nil {
		tx = db.DB
	}

	var fWSectors []FortuneWheelSector
	err := tx.Preload("Benefit").Find(&fWSectors).Error
	if err != nil {
		return fWSectors, logger.WrapError(err, "")
	}

	if len(fWSectors) == 0 {
		return fWSectors, logger.WrapError(err, "")
	}

	for i := range fWSectors {
		if err = fWSectors[i].Benefit.PreloadPolymorphicBenefit(tx); err != nil {
			return fWSectors, logger.WrapError(err, "")
		}
	}

	return fWSectors, nil
}

func WeightedRandomSelection(tx *gorm.DB) (FortuneWheelSector, error) {
	sectors, err := GetAllFortuneWheelSectors(tx)
	if err != nil {
		return FortuneWheelSector{}, logger.WrapError(err, "")
	}

	totalProbability := 0.0
	for i := range sectors {
		totalProbability += sectors[i].Probability
	}

	r := rand.Float64() * totalProbability
	cumulativeProbability := 0.0

	for i := range sectors {
		cumulativeProbability += sectors[i].Probability
		if r < cumulativeProbability {
			return sectors[i], nil
		}
	}

	// This should never happen if probabilities sum to ~1, but just in case:
	return sectors[len(sectors)-1], errors.New("failed to select a sector, falling back to last sector")
}
