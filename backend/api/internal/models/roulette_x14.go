package models

import "time"

type RouletteX14Bet struct {
	ID               int64   `gorm:"primaryKey,autoIncrement"`
	UserID           int64   `gorm:"not null;index"`
	Amount           float64 `gorm:"not null"`
	IsBenefitBet     bool
	FromCashBalance  float64 `json:"-"`
	FromBonusBalance float64 `json:"-"`
	BetColor         string  `gorm:"not null"`
	Outcome          string
	Payout           float64
	CreatedAt        time.Time
}

type RouletteX14GameResult struct {
	ID           int64     `gorm:"primaryKey,autoIncrement"`
	UserID       int64     `gorm:"not null;index"`
	WinningColor string    `gorm:"not null"`
	SectorNumber int       `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
}
