package benefits

const BenefitFortuneWheelType = "benefit_fortune_wheel"

type BenefitFortuneWheel struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	FreeSpinsAmount int
}
