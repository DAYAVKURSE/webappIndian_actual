package requirements

const RequirementClickerType = "requirement_clicker"

type RequirementClicker struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	ClicksAmount int
	TimeDuration int64
	HitLimit     bool
}
