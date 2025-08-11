package service

import (
	"BlessedApi/internal/models/exchange"
	"BlessedApi/pkg/logger"
	"math"
	"math/rand"
	"sync"
	"time"

	"BlessedApi/cmd/db"
	"BlessedApi/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- Config ---
const crashCountdownSec = 10

// Fixed multipliers for creatives (amount -> multiplier)
var CrashFixed = map[int]float64{538: 32, 76: 1.5, 17216: 2.5, 372: 1.5}
var CrashCreativeMode = false // true ⇒ every bet wins 40–65× (off on prod)

// --- Engine commands/events ---

type crashCmd interface{}

type cmdPlaceBet struct {
	Bet  Bet
	Resp chan placeAck
}

type placeAck struct {
	Queued bool
	Err    error
}

type cmdManualCashout struct {
	UserID int64
	Resp   chan error
}

type engineEvent struct {
	Type string
	Data gin.H
}

// --- In-memory bet used by engine runtime ---

type Bet struct {
	UserID            int64
	Amount            float64
	CashOutMultiplier float64 // 0 = manual
	settled           bool
}

// --- Round state ---

type roundState int

const (
	rsCountdown roundState = iota
	rsPlaying
)

type crashEngine struct {
	// deps
	gorm *gorm.DB

	// runtime
	state         roundState
	curRoundID    int64
	curMultiplier float64
	crashPoint    float64
	startAt       time.Time

	betsCurrent []*Bet // bets playing this round
	betsQueue   []*Bet // bets queued for next round

	cmdCh chan crashCmd
	evtCh chan engineEvent

	mu sync.Mutex
}

func newCrashEngine(g *gorm.DB) *crashEngine {
	e := &crashEngine{gorm: g, state: rsCountdown, cmdCh: make(chan crashCmd, 1024), evtCh: make(chan engineEvent, 1024)}
	go e.loop()
	return e
}

func (e *crashEngine) loop() {
	left := crashCountdownSec
	e.emit("countdown_tick", gin.H{"seconds_left": left})
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case cmd := <-e.cmdCh:
			switch c := cmd.(type) {
			case cmdPlaceBet:
				if e.state == rsPlaying { // очередь
					e.betsQueue = append(e.betsQueue, &c.Bet)
					c.Resp <- placeAck{Queued: true}
				} else { // в следующий стартующий раунд
					e.betsCurrent = append(e.betsCurrent, &c.Bet)
					c.Resp <- placeAck{Queued: false}
				}
			case cmdManualCashout:
				err := e.manualCashout(c.UserID)
				c.Resp <- err
			}
		case <-tick.C:
			switch e.state {
			case rsCountdown:
				left--
				if left > 0 {
					e.emit("countdown_tick", gin.H{"seconds_left": left})
					continue
				}
				e.startRound()
				left = crashCountdownSec
			case rsPlaying:
				e.advance()
			}
		}
	}
}

func (e *crashEngine) startRound() {
	e.curRoundID++
	e.curMultiplier = 1
	e.startAt = time.Now()
	// перенесём очередь в текущие
	e.betsCurrent = append(e.betsCurrent, e.betsQueue...)
	e.betsQueue = nil

	// выбрать crash point
	e.crashPoint = e.chooseCrash()
	logger.Info("crashPoint %v", e.crashPoint)
	e.state = rsPlaying
	e.emit("new_round", gin.H{"round_id": e.curRoundID})

	// Persist CrashGame row (created on start)
	game := &models.CrashGame{CrashPointMultiplier: e.crashPoint, StartTime: e.startAt}
	if err := db.DB.Create(game).Error; err == nil {
		// связать активные ставки с этим раундом при необходимости в вашей БД-логике
	}
}

func (e *crashEngine) chooseCrash() float64 {
	if CrashCreativeMode {
		return randF(40, 65)
	}
	for _, b := range e.betsCurrent {
		if m, ok := CrashFixed[int(math.Round(b.Amount))]; ok {
			return m
		}
	}
	return randF(1.2, 10)
}

func (e *crashEngine) advance() {
	// простой прирост (можно заменить на экспоненту)
	e.curMultiplier += 0.1
	e.emit("multiplier", gin.H{"value": e.curMultiplier})

	// авто-кэшаут

	for _, b := range e.betsCurrent {
		if !b.settled && b.CashOutMultiplier > 0 && e.curMultiplier >= b.CashOutMultiplier {
			b.settled = true
			e.emit("cashout", gin.H{"user_id": b.UserID, "win_amount": b.Amount * e.curMultiplier, "multiplier": e.curMultiplier, "is_auto": true})
			logger.Info("cashout ", gin.H{"user_id": b.UserID, "win_amount": b.Amount * e.curMultiplier, "multiplier": e.curMultiplier, "is_auto": true})
			// TODO: тут вызвать вашу экономику (обновление балансов)
			err := crashGameCashout(nil, &models.CrashGameBet{
				UserID:      b.UserID,
				CrashGameID: e.curRoundID,
				Amount:      b.Amount,
			}, e.curMultiplier)
			if err != nil {
				logger.Error(err.Error())
				return
			}

		}
	}

	if e.curMultiplier >= e.crashPoint {
		e.emit("crash", gin.H{"crash_point": e.crashPoint})
		// проигравшие
		for _, b := range e.betsCurrent {
			if !b.settled { /* TODO: пометить проигрыш в БД */
				logger.Info("settle %v", b)
			}
		}
		e.betsCurrent = nil
		e.state = rsCountdown
	}
}

func (e *crashEngine) manualCashout(uid int64) error {
	for _, b := range e.betsCurrent {
		if b.UserID == uid && !b.settled {
			b.settled = true
			e.emit("cashout", gin.H{"user_id": b.UserID, "win_amount": b.Amount * e.curMultiplier, "multiplier": e.curMultiplier, "is_auto": false})
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (e *crashEngine) emit(t string, data gin.H) { e.evtCh <- engineEvent{Type: t, Data: data} }

// last 50 (достаём из БД по факту)
func (e *crashEngine) last50() ([]models.CrashGame, error) {
	var games []models.CrashGame
	err := db.DB.Where("start_time != ? AND end_time != ?", time.Time{}, time.Time{}).
		Order("start_time DESC").Limit(50).Find(&games).Error
	return games, err
}

// helpers
func randF(min, max float64) float64 { return min + rand.Float64()*(max-min) }

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
