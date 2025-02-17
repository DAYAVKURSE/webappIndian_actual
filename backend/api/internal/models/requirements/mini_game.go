package requirements

const RequirementMiniGameType = "requirement_mini_game"

const (
	NvutiGameID    = 1
	DiceGameID     = 2
	RouletteGameID = 3
)

type RequirementMiniGame struct {
	ID          int64 `gorm:"primaryKey;autoIncrement"`
	GameID      int64 // gorm:"index"
	MinBetRupee float64
	BetsAmount  int
	WinsAmount  int
}
