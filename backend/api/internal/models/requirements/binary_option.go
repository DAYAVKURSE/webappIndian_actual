package requirements

const RequirementBinaryOptionType = "requirement_binary_option"

type RequirementBinaryOption struct {
	ID                 int64 `gorm:"primaryKey;autoIncrement"`
	MinBetRupee        float64
	BetsAmount         int
	WinsAmount         int
	TotalWinningsRupee float64
}
