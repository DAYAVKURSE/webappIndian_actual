package requirements

const RequirementTurnoverType = "requirement_turnover"

type RequirementTurnover struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	AmountRupee  float64
	TimeDuration int64
}
