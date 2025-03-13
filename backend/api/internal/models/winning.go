package models

import "time"

type Winning struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null"`
    WinAmount float64   `gorm:"not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
