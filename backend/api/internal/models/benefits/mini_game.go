package benefits

const BenefitMiniGameType = "benefit_mini_game"

type BenefitMiniGame struct {
	ID                  int64 `gorm:"primaryKey;autoIncrement"`
	GameID              int64 // gorm:"index"
	FreeBetsAmount      int
	FreeBetDepositRupee float64
}
