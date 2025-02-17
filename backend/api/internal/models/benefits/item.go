package benefits

import (
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"fmt"

	"gorm.io/gorm"
)

const BenefitItemType = "benefit_item"

type BenefitItem struct {
	ID       int64 `gorm:"primaryKey;autoIncrement"`
	ItemName string
}

// todo: implement
func (benefitItem *BenefitItem) ApplyBenefit(tx *gorm.DB, userID int64) error {
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		return logger.WrapError(err, "Failed to fetch user")
	}

	// Apply the benefit (this is a placeholder, replace with actual benefit logic)
	benefitDescription := fmt.Sprintf("Received benefit: %s", benefitItem.ItemName)

	// Log the benefit application
	logger.BenefitItem("User with ID %d (%s) %s", user.ID, user.Nickname, benefitDescription)

	return nil
}
