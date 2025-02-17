package service

import (
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Exported global instance of the WebSocket service
var RouletteWebsocketService *RouletteX14WebsocketService

//func init() {
//	RouletteWebsocketService = NewRouletteX14WebsocketService()
//}

// RouletteX14WebsocketService handles WebSocket connections for the Roulette X14 game.
type RouletteX14WebsocketService struct {
	connections      map[int64]*websocket.Conn
	mu               sync.Mutex
	lastActivityTime map[int64]time.Time
	betDistribution  map[string]float64
	betMutex         sync.RWMutex
}

func (ws *RouletteX14WebsocketService) updateBetDistribution(bet models.RouletteX14Bet) {
	ws.betMutex.Lock()
	defer ws.betMutex.Unlock()

	ws.betDistribution[bet.BetColor] += bet.Amount

	total := ws.betDistribution["red"] + ws.betDistribution["black"] + ws.betDistribution["green"]

	for color := range ws.betDistribution {
		ws.betDistribution[color] = (ws.betDistribution[color] / total) * 100
	}
}

// NewRouletteX14WebsocketService creates a new instance of RouletteX14WebsocketService.
func NewRouletteX14WebsocketService() *RouletteX14WebsocketService {
	service := &RouletteX14WebsocketService{
		connections:      make(map[int64]*websocket.Conn),
		lastActivityTime: make(map[int64]time.Time),
		betDistribution: map[string]float64{
			"red":   0,
			"black": 0,
			"green": 0,
		},
	}
	go service.cleanupInactiveConnections()
	return service
}

func (ws *RouletteX14WebsocketService) cleanupInactiveConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ws.mu.Lock()
		now := time.Now()
		for userId, lastActivity := range ws.lastActivityTime {
			if now.Sub(lastActivity) > 30*time.Minute {
				if conn, ok := ws.connections[userId]; ok {
					conn.Close()
					delete(ws.connections, userId)
					delete(ws.lastActivityTime, userId)
				}
			}
		}
		ws.mu.Unlock()
	}
}

func SuperviseRouletteX14Game() {
	for {
		logger.Info("Starting Roulette X14 game loop")

		// Run the game loop in a separate goroutine
		done := make(chan bool)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Roulette X14 game loop panicked: %v", r)
					done <- true
				}
			}()

			//StartRouletteX14Game()
		}()

		// Wait for the game loop to finish (which should only happen if there's a panic)
		<-done

		time.Sleep(5 * time.Second)
	}
}

// LiveRouletteX14WebsocketHandler handles the WebSocket connection for live Roulette X14 bets.
func (ws *RouletteX14WebsocketService) LiveRouletteX14WebsocketHandler(c *gin.Context) {
	userId, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("%v", err)
		return
	}

	ws.mu.Lock()
	ws.connections[userId] = conn
	ws.lastActivityTime[userId] = time.Now()
	ws.mu.Unlock()

	defer func() {
		ws.mu.Lock()
		delete(ws.connections, userId)
		delete(ws.lastActivityTime, userId)
		ws.mu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		ws.mu.Lock()
		ws.lastActivityTime[userId] = time.Now()
		ws.mu.Unlock()
	}
}

// BroadcastTimerTick sends the current time remaining until the next spin to all connected WebSocket clients.
func (ws *RouletteX14WebsocketService) BroadcastTimerTick(remainingTime time.Duration, isBettingOpen bool, timerType string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	timerTick := gin.H{
		"type":            "timer_tick",
		"remaining_time":  remainingTime.Seconds(),
		"is_betting_open": isBettingOpen,
		"timer_type":      timerType,
	}

	for _, conn := range ws.connections {
		err := conn.WriteJSON(timerTick)
		if err != nil {
			logger.Error("Failed to broadcast timer tick: %v", err)
			conn.Close()
		}
	}
}

// BroadcastBetToAll sends a user's bet to all connected WebSocket clients.
func (ws *RouletteX14WebsocketService) BroadcastBetToAll(bet models.RouletteX14Bet, user models.User) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.updateBetDistribution(bet)

	betInfo := gin.H{
		"user_id":          user.ID,
		"nickname":         user.Nickname,
		"avatar_id":        user.AvatarID,
		"amount":           bet.Amount,
		"bet_color":        bet.BetColor,
		"bet_distribution": ws.betDistribution,
	}

	for _, conn := range ws.connections {
		err := conn.WriteJSON(betInfo)
		if err != nil {
			logger.Error("Failed to broadcast bet: %v", err)
			conn.Close()
		}
	}
}

func (ws *RouletteX14WebsocketService) resetBetDistribution() {
	ws.betMutex.Lock()
	defer ws.betMutex.Unlock()

	ws.betDistribution = map[string]float64{
		"red":   0,
		"black": 0,
		"green": 0,
	}
}

// SendBetResultToUser sends the result of a bet to the user via WebSocket.
func (ws *RouletteX14WebsocketService) SendBetResultToUser(userId int64, bet models.RouletteX14Bet) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if conn, ok := ws.connections[userId]; ok {
		resultInfo := gin.H{
			"amount":    bet.Amount,
			"bet_color": bet.BetColor,
			"outcome":   bet.Outcome,
			"payout":    bet.Payout,
		}
		err := conn.WriteJSON(resultInfo)
		if err != nil {
			logger.Error("Failed to send result to user %d: %v", userId, err)
			conn.Close()
			delete(ws.connections, userId)
		}
	}
}

// BroadcastSpinResult sends the spin result to all connected WebSocket clients.
func (ws *RouletteX14WebsocketService) BroadcastSpinResult(sector RouletteX14Sector) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	spinResult := gin.H{
		"type":          "spin_result",
		"winning_color": sector.Color,
		"sector_id":     sector.SectorId,
		"sector_number": sector.SectorNumber,
	}

	for _, conn := range ws.connections {
		err := conn.WriteJSON(spinResult)
		if err != nil {
			logger.Error("Failed to broadcast spin result: %v", err)
			conn.Close()
			// You may choose to handle connection removal here
		}
	}
}

// BroadcastNewGameStarting sends a signal to all connected WebSocket clients that a new game is starting.
func (ws *RouletteX14WebsocketService) BroadcastNewGameStarting() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.resetBetDistribution()

	newGameSignal := gin.H{
		"type":    "new_game",
		"message": "New game starting",
	}

	for _, conn := range ws.connections {
		err := conn.WriteJSON(newGameSignal)
		if err != nil {
			logger.Error("Failed to broadcast new game signal: %v", err)
			conn.Close()
			// You may choose to handle connection removal here
		}
	}
}
