package benefits

const BenefitClickerType = "benefit_clicker"

type BenefitClicker struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	TimeDuration    int64
	BonusMultiplier float64
	Reset           bool
}
