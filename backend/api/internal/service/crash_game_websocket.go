package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"errors"
	"math/rand"
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
	gameState        *models.CrashGame
}

var crashPoints = map[int]float64{
	76:     1.5,
	538:    32,
	17216:  2.5,
	372:    1.5,
	1186:   14,
	16604:  4,
	614:    1.5,
	2307:   13,
	29991:  3,
	1476:   1.5,
	5738:   7,
	40166:  3,
	3258:   1.5,
	11629:  4,
	465616: 4.5,
}
func NewCrashGameWebsocketService() *CrashGameWebsocketService {
	service := &CrashGameWebsocketService{
		connections:      make(map[int64]*websocket.Conn),
		lastActivityTime: make(map[int64]time.Time),
		bets:             make(map[int64]*models.CrashGameBet),
		betCount:         0,
		gameState:        &models.CrashGame{},
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

	logger.Info("User %d connected to WebSocket", userId)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed: %v", err)
		return
	}

	w.mu.Lock()
	w.connections[userId] = conn
	w.lastActivityTime[userId] = time.Now()
	w.betCount++
	w.mu.Unlock()

	// Отправляем текущее состояние игры новому клиенту
	if w.gameState != nil && w.gameState.ID != 0 {
		conn.WriteJSON(gin.H{
			"type": "game_state",
			"game": w.gameState,
		})
	}

	defer func() {
		w.mu.Lock()
		delete(w.connections, userId)
		delete(w.lastActivityTime, userId)
		w.mu.Unlock()
		conn.Close()
		logger.Info("User %d disconnected from WebSocket", userId)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		w.mu.Lock()
		w.lastActivityTime[userId] = time.Now()
		w.mu.Unlock()
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

func (w *CrashGameWebsocketService) BroadcastGameStarted() {
	w.mu.Lock()
	defer w.mu.Unlock()

	gameStarted := gin.H{
		"type": "game_started",
	}

	for userId, conn := range w.connections {
		err := conn.WriteJSON(gameStarted)
		if err != nil {
			logger.Error("Failed to broadcast game started to user %d: %v", userId, err)
			conn.Close()
			delete(w.connections, userId)
			delete(w.lastActivityTime, userId)
		}
	}
}

func (w *CrashGameWebsocketService) BroadcastGameCrash(crashPoint float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	crashEvent := gin.H{
		"type":        "game_crash",
		"crash_point": crashPoint,
	}

	for userId, conn := range w.connections {
		err := conn.WriteJSON(crashEvent)
		if err != nil {
			logger.Error("Failed to broadcast crash event to user %d: %v", userId, err)
			conn.Close()
			delete(w.connections, userId)
			delete(w.lastActivityTime, userId)
		}
	}
}

func (w *CrashGameWebsocketService) SendMultiplierToUser(currentGame *models.CrashGame) {
	w.mu.Lock()
	w.gameState = currentGame
	w.mu.Unlock()

	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Создаем копию подключений для безопасной отправки
	w.mu.Lock()
	connections := make(map[int64]*websocket.Conn)
	for userId, conn := range w.connections {
		connections[userId] = conn
	}
	w.mu.Unlock()

	if len(connections) == 0 {
		return
	}

	startTime := time.Now()
	baseMultiplier := 1.0

	// Отправляем обновление множителя каждые 16мс (примерно 60 FPS)
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		elapsed := time.Since(startTime).Seconds()
		
		// Базовое увеличение множителя
		baseMultiplier = 1.0 + (elapsed * 0.1)
		
		// Добавляем случайную составляющую для более естественного движения
		randomFactor := 1.0 + (rand.Float64() * 0.01)
		currentMultiplier := baseMultiplier * randomFactor
		
		multiplierUpdate := gin.H{
			"type":      "multiplier_update",
			"multiplier": currentMultiplier,
			"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
			"elapsed":   elapsed,
		}

		w.mu.Lock()
		for userId, conn := range connections {
			err := conn.WriteJSON(multiplierUpdate)
			if err != nil {
				logger.Error("Failed to send multiplier update to user %d: %v", userId, err)
				conn.Close()
				delete(connections, userId)
				delete(w.connections, userId)
				delete(w.lastActivityTime, userId)
			}
		}
		w.mu.Unlock()

		// Проверяем, достиг ли множитель точки краша
		if currentMultiplier >= currentGame.CrashPointMultiplier {
			w.BroadcastGameCrash(currentGame.CrashPointMultiplier)
			break
		}
	}
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
func (ws *RouletteX14WebsocketService) SendCrashGameBetResultToUser(userId int64, bet models.CrashGameBet) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if conn, ok := ws.connections[userId]; ok {
		resultInfo := gin.H{
			"bet_amount":          bet.Amount,
			"win_amount":          bet.WinAmount,
			"cash_out_multiplier": bet.CashOutMultiplier,
			"status":              bet.Status,
		}
		err := conn.WriteJSON(resultInfo)
		if err != nil {
			logger.Error("Failed to send bet result to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
		}
	}
}

// BroadcastCrashGameBetToAll sends a user's bet to all connected WebSocket clients.
// func (ws *RouletteX14WebsocketService) BroadcastCrashGameBetToAll(bet models.CrashGameBet, user models.User) {
// 	ws.mu.Lock()
// 	defer ws.mu.Unlock()

// 	betInfo := gin.H{
// 		"user_id":             user.ID,
// 		"nickname":            user.Nickname,
// 		"avatar_id":           user.AvatarID,
// 		"amount":              bet.Amount,
// 		"cash_out_multiplier": bet.CashOutMultiplier,
// 	}

// 	for _, conn := range ws.connections {
// 		err := conn.WriteJSON(betInfo)
// 		if err != nil {
// 			logger.Error("Failed to broadcast bet: %v", err)
// 			conn.Close()
// 		}
// 	}
// }
