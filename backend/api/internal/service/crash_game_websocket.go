package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"errors"
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
	
	// Еще раз проверяем значение в базе данных
	var gameFromDB models.CrashGame
	if err := db.DB.First(&gameFromDB, currentGame.ID).Error; err != nil {
		logger.Error("Failed to re-read game %d before sending multipliers: %v", currentGame.ID, err)
	} else {
		logger.Info("Game %d crash point from DB: %.2f", gameFromDB.ID, gameFromDB.CrashPointMultiplier)
		
		// Если значение в базе отличается от значения в памяти, обновляем
		if gameFromDB.CrashPointMultiplier != currentGame.CrashPointMultiplier {
			logger.Info("Updating crash point in memory: %.2f -> %.2f", 
				currentGame.CrashPointMultiplier, gameFromDB.CrashPointMultiplier)
			currentGame.CrashPointMultiplier = gameFromDB.CrashPointMultiplier
		}
	}
	
	ws.mu.Lock()

	var currentMultiplier float64
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
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		currentMultiplier = currentGame.CalculateMultiplier()
		
		// Сглаживание множителя (экспоненциальное усреднение)
		smoothedMultiplier := (lastSentMultiplier*0.8 + currentMultiplier*0.2)
		
		// Проверяем, достигли ли мы точки краша
		if smoothedMultiplier >= currentGame.CrashPointMultiplier {
			logger.Info("Game %d reached crash point: %.2f >= %.2f", 
				currentGame.ID, smoothedMultiplier, currentGame.CrashPointMultiplier)
			crashPointReached = true
			ws.BroadcastGameCrash(currentGame.CrashPointMultiplier)
			break
		}

		// Отправляем множитель только если он изменился достаточно
		if math.Abs(smoothedMultiplier-lastSentMultiplier) > 0.01 {
			multiplierInfo := gin.H{
				"type":       "multiplier_update",
				"multiplier": smoothedMultiplier,
				"timestamp":  time.Now().UnixNano() / int64(time.Millisecond),
				"elapsed":    time.Since(startTime).Seconds(),
			}
			
			ws.mu.Lock()
			for userId, conn := range connections {
				// Проверяем, есть ли активная ставка у пользователя
				if bet, exists := ws.bets[userId]; exists && bet.Status == "active" {
					// Проверяем автокэшаут
					if bet.CashOutMultiplier > 0 && smoothedMultiplier >= bet.CashOutMultiplier {
						logger.Info("Auto cashout for user %d at %.2fx", userId, smoothedMultiplier)
						// Обрабатываем автокэшаут
						if err := crashGameCashout(nil, bet, smoothedMultiplier); err != nil {
							logger.Error("Unable to auto cashout for user %d: %v", userId, err)
							continue
						}
						ws.ProcessCashout(userId, smoothedMultiplier, true)
						continue
					}

					// Отправляем обновление множителя
					err := conn.WriteJSON(multiplierInfo)
					if err != nil {
						logger.Error("Failed to send multiplier to user %d: %v", userId, err)
						// Пытаемся обработать ошибку более грациозно
						if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
							// Только если это неожиданная ошибка закрытия, закрываем соединение
							conn.Close()
							delete(connections, userId)
							delete(ws.connections, userId)
						}
					}
				} else {
					// Для пользователей без активных ставок тоже отправляем множитель
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
		}
	}

	// Если достигли точки краша, обрабатываем все ставки
	if crashPointReached {
		logger.Info("Game %d crashed, processing all active bets", currentGame.ID)
		ws.mu.Lock()
		for userId, bet := range ws.bets {
			if bet.Status == "active" {
				logger.Info("Marking bet as lost for user %d", userId)
				// Обновляем статус ставки
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
		"userId":             userId,
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
