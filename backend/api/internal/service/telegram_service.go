package service

import (
    "BlessedApi/pkg/logger"
    "fmt"
    "net/http"
    "time"
)

const (
    telegramBotToken = "YOUR_BOT_TOKEN" // Замените на ваш токен бота
    managerChatID    = "YOUR_CHAT_ID"   // Замените на ID чата менеджера
    telegramAPIURL   = "https://api.telegram.org/bot%s/sendMessage"
)

// SendTelegramMessage отправляет сообщение в Telegram
func SendTelegramMessage(message string) error {
    url := fmt.Sprintf(telegramAPIURL, telegramBotToken)
    
    payload := map[string]interface{}{
        "chat_id": managerChatID,
        "text":    message,
        "parse_mode": "HTML",
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal telegram message: %v", err)
    }

    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to send telegram message: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("telegram API returned non-200 status: %d", resp.StatusCode)
    }

    return nil
} 