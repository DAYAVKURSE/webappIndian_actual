package models

import "time"

type Winning struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    int64     `gorm:"not null"`  // Изменили uint на int64
    WinAmount float64   `gorm:"not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
