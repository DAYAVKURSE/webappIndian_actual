package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"BlessedApi/internal/models/exchange"
	"BlessedApi/pkg/logger"
	"encoding/json"
	"errors"
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
	errInsufficientBalance := errors.New("insufficient balance")
	errExistingBet := errors.New("active bet already exists")

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
	if in.Amount <= 0 {
		c.JSON(400, gin.H{"error": "Amount must be > 0"})
		return
	}

	// ТРАНЗАКЦИЯ с проверкой существующей активной ставки
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		// запрет второй активной ставки
		var existing models.CrashGameBet
		if err := tx.Where("user_id = ? AND status = ?", uid, "active").
			Order("id DESC").First(&existing).Error; err == nil {
			return errExistingBet
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return logger.WrapError(err, "check existing bet failed")
		}

		var user models.User
		if err := tx.First(&user, uid).Error; err != nil {
			return logger.WrapError(err, "load user failed")
		}

		bonusBalance, err := exchange.GetUserExchangedBalanceAmount(tx, user.ID)
		if err != nil {
			return logger.WrapError(err, "load bonus balance failed")
		}
		if user.BalanceRupee+bonusBalance < in.Amount {
			return errInsufficientBalance
		}

		fromCash, fromBonus, err := exchange.UseExchangeBalancePayment(tx, &user, in.Amount)
		if err != nil {
			return logger.WrapError(err, "debit failed")
		}

		bet := models.CrashGameBet{
			UserID:            uid,
			Amount:            fromCash + fromBonus,
			FromCashBalance:   fromCash,
			FromBonusBalance:  fromBonus,
			CashOutMultiplier: in.CashOutMultiplier,
			Status:            "active",
		}
		if err := tx.Create(&bet).Error; err != nil {
			return logger.WrapError(err, "create bet failed")
		}
		return nil
	}); err != nil {
		switch {
		case errors.Is(err, errInsufficientBalance):
			c.JSON(402, gin.H{"error": "Insufficient balance"})
		case errors.Is(err, errExistingBet):
			c.JSON(400, gin.H{"error": "You already have an active bet"})
		default:
			logger.Error("Place bet failed: %v", err)
			c.JSON(500, gin.H{"error": "Failed to place bet"})
		}
		return
	}

	// только после успешной транзакции — в движок
	ackCh := make(chan placeAck, 1)
	CrashGameWS.eng.cmdCh <- cmdPlaceBet{
		Bet:  Bet{UserID: uid, Amount: in.Amount, CashOutMultiplier: in.CashOutMultiplier},
		Resp: ackCh,
	}
	ack := <-ackCh
	if ack.Err != nil {
		c.JSON(500, gin.H{"error": "Engine rejected bet"})
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
		switch err.Error() {
		case "no active round":
			c.JSON(400, gin.H{"error": "no active round"})
		default:
			// gorm.ErrRecordNotFound и прочие
			c.JSON(404, gin.H{"error": "no active bet"})
		}
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
