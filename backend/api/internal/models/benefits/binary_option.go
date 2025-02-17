package benefits

const BenefitBinaryOptionType = "benefit_binary_option"

type BenefitBinaryOption struct {
	ID                  int64 `gorm:"primaryKey;autoIncrement"`
	FreeBetsAmount      int
	FreeBetDepositRupee float64
}
