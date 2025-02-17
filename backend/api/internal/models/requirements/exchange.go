package requirements

const RequirementExchangeType = "requirement_exchange"

type RequirementExchange struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	BCoinsAmount float64
}
