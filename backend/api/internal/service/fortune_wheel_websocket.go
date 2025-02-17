package service

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"BlessedApi/pkg/logger"
	"BlessedApi/pkg/redis"
)

// FortuneWheelWebsocketService is responsible for handling WebSocket connections Fortune Wheel wins.
type FortuneWheelWebsocketService struct {
	redisService *redis.RedisService
}

// NewFortuneWheelWebsocketService creates a new instance of FortuneWheelWebsocketService.
func NewFortuneWheelWebsocketService(redisService *redis.RedisService) *FortuneWheelWebsocketService {
	return &FortuneWheelWebsocketService{
		redisService: redisService,
	}
}

// GetRecentWins handles GET requests to fetch recent Fortune Wheel wins.
func (f *FortuneWheelWebsocketService) GetRecentWins(c *gin.Context) {
	wins, err := f.fetchRecentWins(c.Request.Context(), 10) // Fetch last 10 wins
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}
	if len(wins) < 1 {
		c.String(404, "[]")
		return
	}
	c.JSON(200, wins)
}

// LiveWinsWebsocketHandler handles the WebSocket connection for live Fortune Wheel wins.
func (f *FortuneWheelWebsocketService) LiveWinsWebsocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("%v", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastWinTimestamp int64

	for range ticker.C { // Continuously fetch and send the latest win data
		wins, err := f.fetchRecentWins(c.Request.Context(), 1) // Fetch only the latest win
		if err != nil {
			logger.Error("%v", err)
			return
		}

		if len(wins) > 0 {
			latestWin := wins[0]
			if latestWin.Timestamp > lastWinTimestamp { // Send only if the latest win is newer
				if err := conn.WriteJSON(latestWin); err != nil {
					logger.Error("%v", err)
					return
				}
				lastWinTimestamp = latestWin.Timestamp
			}
		}
	}
}

// fetchRecentWins retrieves recent Fortune Wheel wins from Redis.
func (f *FortuneWheelWebsocketService) fetchRecentWins(ctx context.Context, limit int) ([]FortuneWheelWinData, error) {
	keys, err := f.fetchSortedKeys(ctx)
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	if len(keys) > limit {
		keys = keys[len(keys)-limit:]
	}

	return f.fetchWinData(ctx, keys)
}

// fetchSortedKeys retrieves and sorts all Fortune Wheel win keys from Redis.
func (f *FortuneWheelWebsocketService) fetchSortedKeys(ctx context.Context) ([]string, error) {
	keys, err := f.redisService.Client().Keys(ctx, "fortune_wheel:win:*").Result()
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	sort.Strings(keys)
	return keys, nil
}

// fetchWinData fetches the win data for the given keys from Redis.
func (f *FortuneWheelWebsocketService) fetchWinData(ctx context.Context, keys []string) ([]FortuneWheelWinData, error) {
	var winData []FortuneWheelWinData

	for _, key := range keys {
		data, err := f.redisService.GetKey(ctx, key)
		if err != nil {
			return nil, logger.WrapError(err, "")
		}

		var win FortuneWheelWinData
		if err := json.Unmarshal([]byte(data), &win); err != nil {
			return nil, logger.WrapError(err, "")
		}

		winData = append(winData, win)
	}

	return winData, nil
}
