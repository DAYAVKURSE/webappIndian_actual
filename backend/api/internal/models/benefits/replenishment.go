package benefits

const BenefitReplenishmentType = "benefit_replenishment"

type BenefitReplenishment struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	BonusMultiplier float64
	TimeDuration    int64
}
