package models

import "time"

type Winning struct {
    ID        int64     `json:"id" gorm:"primaryKey"`
    UserID    int64     `json:"user_id"`
    Amount    float64   `json:"amount"`
    CreatedAt time.Time `json:"created_at"`
}
