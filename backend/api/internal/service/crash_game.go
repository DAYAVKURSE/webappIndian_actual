package service

import (
	"BlessedApi/internal/models/exchange"
	"BlessedApi/pkg/logger"
	"errors"
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
				handled := false
				if e.state == rsPlaying {
					for _, qb := range e.betsQueue {
						if qb.UserID == c.Bet.UserID && !qb.settled {
							c.Resp <- placeAck{Queued: true, Err: errors.New("active bet already exists")}
							handled = true
							break
						}
					}
					if !handled {
						e.betsQueue = append(e.betsQueue, &c.Bet)
						c.Resp <- placeAck{Queued: true}
					}
				} else {
					for _, cb := range e.betsCurrent {
						if cb.UserID == c.Bet.UserID && !cb.settled {
							c.Resp <- placeAck{Queued: false, Err: errors.New("active bet already exists")}
							handled = true
							break
						}
					}
					if !handled {
						e.betsCurrent = append(e.betsCurrent, &c.Bet)
						c.Resp <- placeAck{Queued: false}
					}
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
	e.emit("multiplier_update", gin.H{"value": e.curMultiplier})

	// авто-кэшаут

	for _, b := range e.betsCurrent {
		if !b.settled && b.CashOutMultiplier > 0 && e.curMultiplier >= b.CashOutMultiplier {
			b.settled = true
			e.emit("cashout", gin.H{"user_id": b.UserID, "win_amount": b.Amount * e.curMultiplier, "multiplier": e.curMultiplier, "is_auto": true})
			logger.Info("cashout ", gin.H{"user_id": b.UserID, "win_amount": b.Amount * e.curMultiplier, "multiplier": e.curMultiplier, "is_auto": true})
			// внутри e.advance() при автокэшауте:
			err := crashGameCashout(nil, &models.CrashGameBet{
				UserID:      b.UserID,
				CrashGameID: e.curRoundID, // <— привязываем к текущему раунду
			}, e.curMultiplier)

			if err != nil {
				logger.Error(err.Error())
				return
			}

		}
	}

	if e.curMultiplier >= e.crashPoint {
		e.emit("crash", gin.H{"crash_point": e.crashPoint})

		// проигравшие — обновляем БД
		for _, b := range e.betsCurrent {
			if b.settled {
				continue
			}
			// find latest active bet of this user and mark lost
			_ = db.DB.Transaction(func(tx *gorm.DB) error {
				var dbBet models.CrashGameBet
				if err := tx.Where("user_id = ? AND status = ?", b.UserID, "active").
					Order("id DESC").
					First(&dbBet).Error; err != nil {
					return nil // если нет — просто пропустим
				}
				dbBet.Status = "lost"
				dbBet.CrashGameID = e.curRoundID
				return tx.Save(&dbBet).Error
			})
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

// Bet must exist: find user's latest active bet and settle it as win.
func crashGameCashout(tx *gorm.DB, bet *models.CrashGameBet, currentMultiplier float64) error {
	if tx == nil {
		tx = db.DB
	}

	// Найти активную ставку пользователя (последнюю)
	var dbBet models.CrashGameBet
	if err := tx.
		Where("user_id = ? AND status = ?", bet.UserID, "active").
		Order("id DESC").
		First(&dbBet).Error; err != nil {
		return logger.WrapError(err, "active bet not found for cashout")
	}

	// Проставим CrashGameID, если движок его знает (ты его передаёшь из engine)
	if bet.CrashGameID != 0 {
		dbBet.CrashGameID = bet.CrashGameID
	}

	dbBet.Status = "won"
	dbBet.CashOutMultiplier = currentMultiplier
	dbBet.WinAmount = dbBet.Amount * currentMultiplier

	if err := tx.Save(&dbBet).Error; err != nil {
		return logger.WrapError(err, "failed to update bet")
	}

	var user models.User
	if err := tx.First(&user, dbBet.UserID).Error; err != nil {
		return logger.WrapError(err, "failed to fetch user")
	}

	// Начисления: пропорционально источникам
	toCashBalance := dbBet.FromCashBalance * currentMultiplier
	toBonusBalance := dbBet.FromBonusBalance * currentMultiplier

	win := models.Winning{
		UserID:    user.ID,
		WinAmount: toCashBalance + toBonusBalance,
	}
	if err := tx.Create(&win).Error; err != nil {
		return logger.WrapError(err, "Failed to record winning")
	}

	if err := exchange.UpdateUserBalances(tx, &user, toCashBalance, toBonusBalance, false); err != nil {
		return logger.WrapError(err, "failed to update user balances")
	}

	return nil
}
