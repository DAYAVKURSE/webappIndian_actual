package service

import (
    "BlessedApi/internal/middleware"
    "BlessedApi/pkg/logger"
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

const (
    paymentAPIURL = "https://pay-crm.com/Remotes/create-payment-page"
    webhookID     = "abc1234"
    returnURL     = "https://blessed.one/return"
)

type PaymentRequest struct {
    Amount         int      `json:"amount"`
    Currency       string   `json:"currency"`
    Buttons        []int    `json:"buttons"`
    PaymentSystem  []string `json:"payment_system"`
    CustomUserID   string   `json:"custom_user_id"`
    ReturnURL      string   `json:"return_url"`
    Language       string   `json:"language"`
    WebhookID      string   `json:"webhook_id"`
}

type PaymentResponse struct {
    Success bool   `json:"success"`
    URL     string `json:"url"`
    OrderID string `json:"order_id"`
}

func CreatePaymentPage(c *gin.Context) {
    userID, err := middleware.GetUserIDFromGinContext(c)
    if err != nil {
        logger.Error("Failed to get user ID: %v", err)
        c.Status(500)
        return
    }

    var input struct {
        Amount int `json:"amount" binding:"required,min=1000"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "Invalid input"})
        return
    }

    // Создаем запрос к платежной системе
    paymentReq := PaymentRequest{
        Amount:        input.Amount,
        Currency:      "INR",
        Buttons:       []int{1000, 2000, 5000},
        PaymentSystem: []string{"paytm", "phonepe", "upi_p2p"},
        CustomUserID:  fmt.Sprintf("user_%d", userID),
        ReturnURL:     fmt.Sprintf("%s?user_id=user_%d", returnURL, userID),
        Language:      "EN",
        WebhookID:     webhookID,
    }

    // Конвертируем запрос в JSON
    jsonData, err := json.Marshal(paymentReq)
    if err != nil {
        logger.Error("Failed to marshal payment request: %v", err)
        c.Status(500)
        return
    }

    // Создаем HTTP запрос
    req, err := http.NewRequest("POST", paymentAPIURL, bytes.NewBuffer(jsonData))
    if err != nil {
        logger.Error("Failed to create request: %v", err)
        c.Status(500)
        return
    }

    req.Header.Set("Content-Type", "application/json")

    // Отправляем запрос
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    resp, err := client.Do(req)
    if err != nil {
        logger.Error("Failed to send request: %v", err)
        c.Status(500)
        return
    }
    defer resp.Body.Close()

    // Читаем ответ
    var paymentResp PaymentResponse
    if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
        logger.Error("Failed to decode response: %v", err)
        c.Status(500)
        return
    }

    if !paymentResp.Success {
        logger.Error("Payment creation failed")
        c.Status(500)
        return
    }

    // Возвращаем URL для перенаправления
    c.JSON(200, gin.H{
        "url": paymentResp.URL,
    })
} 