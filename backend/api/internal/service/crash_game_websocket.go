package service

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"BlessedApi/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// CrashGameWS — совместимо с вашими роутами
var CrashGameWS *CrashHub

func InitCrashGameModule(gormDB *gorm.DB) {
	CrashGameWS = NewCrashHub(gormDB)
}

type CrashHub struct {
	upg   websocket.Upgrader
	mu    sync.Mutex
	conns map[int64]*websocket.Conn

	eng *crashEngine
}

func NewCrashHub(g *gorm.DB) *CrashHub {
	h := &CrashHub{
		upg:   websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		conns: map[int64]*websocket.Conn{},
	}
	h.eng = newCrashEngine(g)
	go h.fanout()
	return h
}

// =============== WS ===============
func (h *CrashHub) LiveCrashGameWebsocketHandler(c *gin.Context) {
	uid, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		c.Status(401)
		return
	}
	conn, err := h.upg.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	if old, ok := h.conns[uid]; ok {
		old.Close()
	}
	h.conns[uid] = conn
	h.mu.Unlock()

	conn.WriteJSON(gin.H{"type": "connection_success", "message": "Connected"})

	go func() { // reader to detect close
		for {
			if _, _, e := conn.ReadMessage(); e != nil {
				break
			}
		}
		h.mu.Lock()
		delete(h.conns, uid)
		h.mu.Unlock()
		conn.Close()
	}()
}

func (h *CrashHub) fanout() {
	for ev := range h.eng.evtCh {
		payload, _ := json.Marshal(gin.H{"type": ev.Type, "ts": time.Now().UnixMilli(), "data": ev.Data})
		h.mu.Lock()
		for uid, c := range h.conns {
			if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
				c.Close()
				delete(h.conns, uid)
			}
		}
		h.mu.Unlock()
	}
}

// =============== HTTP ===============

type crashPlaceInput struct {
	Amount            float64 `json:"Amount"`
	CashOutMultiplier float64 `json:"CashOutMultiplier"`
}

func PlaceCrashGameBet(c *gin.Context) {
	// оставляем сигнатуру
	uid, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		c.Status(401)
		return
	}
	var in crashPlaceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	ackCh := make(chan placeAck, 1)

	CrashGameWS.eng.cmdCh <- cmdPlaceBet{Bet: Bet{UserID: uid, Amount: in.Amount, CashOutMultiplier: in.CashOutMultiplier}, Resp: ackCh}
	ack := <-ackCh

	if ack.Err != nil {
		c.JSON(400, gin.H{"error": ack.Err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "accepted", "queued": ack.Queued})
}

func ManualCashout(c *gin.Context) {
	uid, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		c.Status(401)
		return
	}
	rc := make(chan error, 1)
	CrashGameWS.eng.cmdCh <- cmdManualCashout{UserID: uid, Resp: rc}
	if err := <-rc; err != nil {
		c.JSON(404, gin.H{"error": "no active bet"})
		return
	}
	c.JSON(200, gin.H{"status": "cashed_out"})
}

func (h *CrashHub) GetLast50CrashGames(c *gin.Context) {
	games, err := h.eng.last50()
	if err != nil {
		c.Status(500)
		return
	}
	c.JSON(200, gin.H{"results": games})
}
