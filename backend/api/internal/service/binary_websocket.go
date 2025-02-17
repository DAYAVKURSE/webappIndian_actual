package service

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"BlessedApi/pkg/binance"
	"BlessedApi/pkg/logger"
	"BlessedApi/pkg/redis"
)

// upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// APIWebsocketService is responsible for handling WebSocket connections and data processing.
type APIWebsocketServiceBinaryOptions struct {
	redisService *redis.RedisService              // Redis service for fetching and storing kline data.
	binanceWS    *binance.BinanceWebsocketService // Binance WebSocket service (not used in the current code).
}

// NewAPIWebsocketService creates a new instance of APIWebsocketService.
func NewAPIWebsocketServiceBinaryOptions(redisService *redis.RedisService, binanceWS *binance.BinanceWebsocketService) *APIWebsocketServiceBinaryOptions {
	return &APIWebsocketServiceBinaryOptions{
		redisService: redisService,
		binanceWS:    binanceWS,
	}
}

// LatestKlineWebsocketHandler handles WebSocket connection and sends only the latest kline data.
func (a *APIWebsocketServiceBinaryOptions) LatestKlineWebsocketHandler(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("%v", err)
		return
	}
	defer conn.Close()

	// Create a ticker to trigger data updates every second.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Continuously fetch and send the latest kline data on each tick.
	for range ticker.C {
		// Fetch the latest kline data.
		latestKline, err := a.getLatestKline(c.Request.Context())
		if err != nil {
			logger.Error("%v", err)
			return
		}

		// Send the latest kline data to the client.
		if err := conn.WriteJSON(latestKline); err != nil {
			// There was logs, now there is no logs
			return
		}
	}
}

// WebsocketHandler handles the WebSocket connection, sending initial and periodic kline data.
func (a *APIWebsocketServiceBinaryOptions) WebsocketHandler(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("%v", err)
		return
	}
	defer conn.Close()

	// Fetch the initial kline data window.
	data, err := a.getKlineDataWindow(c.Request.Context(), 300)
	if err != nil {
		logger.Error("%v", err)
		return
	}

	// Send the initial kline data to the client.
	if err := conn.WriteJSON(data); err != nil {
		logger.Error("%v", err)
		return
	}

	// Create a ticker to trigger data updates every second.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Continuously fetch and send the latest kline data on each tick.
	for range ticker.C {
		// Fetch the latest kline data.
		latestKline, err := a.getLatestKline(c.Request.Context())
		if err != nil {
			logger.Error("%v", err)
			return
		}

		// Update the data window with the latest kline data.
		data = updateDataWindow(data, latestKline, 300)

		// Send the updated kline data to the client.
		if err := conn.WriteJSON(data); err != nil {
			// There was logs, now there is no logs
			return
		}
	}
}

// getKlineDataWindow retrieves a window of kline data from Redis, up to the specified size.
func (a *APIWebsocketServiceBinaryOptions) getKlineDataWindow(ctx context.Context, windowSize int) ([]binance.KlineData, error) {
	// Fetch and sort keys from Redis.
	keys, err := a.fetchSortedKeys(ctx)
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	// Trim the keys to the window size if necessary.
	if len(keys) > windowSize {
		keys = keys[len(keys)-windowSize:]
	}

	// Fetch the kline data for the given keys.
	return a.fetchKlineData(ctx, keys)
}

// getLatestKline retrieves the latest kline data from Redis.
func (a *APIWebsocketServiceBinaryOptions) getLatestKline(ctx context.Context) (binance.KlineData, error) {
	// Fetch and sort keys from Redis.
	keys, err := a.fetchSortedKeys(ctx)
	if err != nil || len(keys) == 0 {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	// Fetch the kline data for the most recent key.
	klineData, err := a.fetchSingleKlineData(ctx, keys[len(keys)-1])
	if err != nil {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	return klineData, nil
}

// fetchSortedKeys retrieves and sorts all kline data keys from Redis.
func (a *APIWebsocketServiceBinaryOptions) fetchSortedKeys(ctx context.Context) ([]string, error) {
	// Fetch keys matching the kline pattern.
	keys, err := a.redisService.Client().Keys(ctx, "binance_kline_*").Result()
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	// Sort the keys to ensure chronological order.
	sort.Strings(keys)
	return keys, nil
}

// fetchKlineData fetches the kline data for the given keys from Redis.
func (a *APIWebsocketServiceBinaryOptions) fetchKlineData(ctx context.Context, keys []string) ([]binance.KlineData, error) {
	var klineData []binance.KlineData

	// Iterate through each key and fetch the corresponding kline data.
	for _, key := range keys {
		data, err := a.redisService.GetKey(ctx, key)
		if err != nil {
			return nil, logger.WrapError(err, "")
		}

		var kline binance.KlineData
		if err := json.Unmarshal([]byte(data), &kline); err != nil {
			return nil, logger.WrapError(err, "")
		}

		// Append the unmarshaled kline data to the result.
		klineData = append(klineData, kline)
	}

	return klineData, nil
}

// fetchSingleKlineData fetches a single kline data entry from Redis using the provided key.
func (a *APIWebsocketServiceBinaryOptions) fetchSingleKlineData(ctx context.Context, key string) (binance.KlineData, error) {
	data, err := a.redisService.GetKey(ctx, key)
	if err != nil {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	var kline binance.KlineData
	if err := json.Unmarshal([]byte(data), &kline); err != nil {
		return binance.KlineData{}, logger.WrapError(err, "")
	}

	return kline, nil
}

// updateDataWindow updates the data window by adding the latest kline data
// and removing the oldest data if the window exceeds the maximum size.
func updateDataWindow(data []binance.KlineData, latestKline binance.KlineData, maxSize int) []binance.KlineData {
	if len(data) >= maxSize {
		// Remove the oldest data to maintain the window size.
		data = data[1:]
	}
	// Add the latest kline data to the window.
	return append(data, latestKline)
}
