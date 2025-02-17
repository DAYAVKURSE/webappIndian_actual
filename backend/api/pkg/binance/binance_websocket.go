package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"BlessedApi/pkg/logger"
	"BlessedApi/pkg/redis"
)

type KlineData struct {
	OpenTime                 int64   `json:"openTime"`
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	Volume                   float64 `json:"volume"`
	CloseTime                int64   `json:"closeTime"`
	QuoteAssetVolume         float64 `json:"quoteAssetVolume"`
	NumberOfTrades           int64   `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  float64 `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume float64 `json:"takerBuyQuoteAssetVolume"`
}

type BinanceWebsocketService struct {
	redisService *redis.RedisService
	wsConn       *websocket.Conn
}

func NewBinanceWebsocketService(redisService *redis.RedisService) *BinanceWebsocketService {
	return &BinanceWebsocketService{
		redisService: redisService,
	}
}

func (b *BinanceWebsocketService) Start() {
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@kline_1s"}
	logger.Info("Connecting to Binance WebSocket at %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Fatal("%v", err)
	}

	b.wsConn = conn
	logger.Info("Connected to Binance WebSocket.")

	// Set ping/pong handlers
	b.setupPingPongHandlers()

	go b.readMessages()
}

func (b *BinanceWebsocketService) setupPingPongHandlers() {
	// Handle incoming ping frames by replying with a pong frame
	b.wsConn.SetPingHandler(func(appData string) error {
		err := b.wsConn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second*10))
		if err != nil {
			logger.Error("%v", err)
			return logger.WrapError(err, "")
		}
		return nil
	})
}

func (b *BinanceWebsocketService) readMessages() {
	for {
		_, message, err := b.wsConn.ReadMessage()
		if err != nil {
			logger.Error("%v", err)
			return
		}

		var event map[string]interface{}
		if err := json.Unmarshal(message, &event); err != nil {
			logger.Error("%v", err)
			continue
		}

		err = b.handleKlineEvent(event)
		if err != nil {
			logger.Error("%v", err)
			continue
		}
	}
}

func (b *BinanceWebsocketService) handleKlineEvent(event map[string]interface{}) error {
	kline, ok := event["k"].(map[string]interface{})
	if !ok {
		return logger.WrapError(fmt.Errorf("unable to cast kline: %v", event["k"]), "")
	}

	openTime := int64(kline["t"].(float64))
	closeTime := int64(kline["T"].(float64))

	open, err := strconv.ParseFloat(kline["o"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	high, err := strconv.ParseFloat(kline["h"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	low, err := strconv.ParseFloat(kline["l"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	close, err := strconv.ParseFloat(kline["c"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	volume, err := strconv.ParseFloat(kline["v"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	quoteAssetVolume, err := strconv.ParseFloat(kline["q"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	numberOfTrades := int64(kline["n"].(float64))

	takerBuyBaseAssetVolume, err := strconv.ParseFloat(kline["V"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	takerBuyQuoteAssetVolume, err := strconv.ParseFloat(kline["Q"].(string), 64)
	if err != nil {
		return logger.WrapError(err, "")
	}

	data := KlineData{
		OpenTime:                 openTime,
		Open:                     open,
		High:                     high,
		Low:                      low,
		Close:                    close,
		Volume:                   volume,
		CloseTime:                closeTime,
		QuoteAssetVolume:         quoteAssetVolume,
		NumberOfTrades:           numberOfTrades,
		TakerBuyBaseAssetVolume:  takerBuyBaseAssetVolume,
		TakerBuyQuoteAssetVolume: takerBuyQuoteAssetVolume,
	}

	err = b.storeKlineData(data)
	if err != nil {
		return logger.WrapError(err, "")
	}
	return nil
}

func (b *BinanceWebsocketService) storeKlineData(data KlineData) error {
	ctx := context.Background()
	key := fmt.Sprintf("binance_kline_%d", data.OpenTime)
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return logger.WrapError(err, "")
	}

	// Store the new Kline data
	err = b.redisService.SetKey(ctx, key, dataBytes, 5*time.Minute)
	if err != nil {
		return logger.WrapError(err, "")
	}

	// Fetch all keys matching the kline data pattern to get the count
	keys, err := b.redisService.Client().Keys(ctx, "binance_kline_*").Result()
	if err != nil {
		return logger.WrapError(err, "")
	}

	// Calculate the timestamp 5 minutes ago
	cutoffTime := data.OpenTime - (5 * time.Minute).Milliseconds()

	// Remove any data older than 5 minutes
	for _, k := range keys {
		// Extract timestamp from key
		tsStr := k[len("binance_kline_"):]
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			logger.Warn("Unable to extract timestamp from key %v", err)
			continue
		}

		// If the data is older than 5 minutes, delete the key
		if ts < cutoffTime {
			err := b.redisService.Client().Del(ctx, k).Err()
			if err != nil {
				return logger.WrapError(err, "")
			}
		}
	}
	return nil
}
