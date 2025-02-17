package requirements

const RequirementReplenishmentType = "requirement_replenishment"

type RequirementReplenishment struct {
	ID          int64 `gorm:"primaryKey;autoIncrement"`
	AmountRupee float64
}
