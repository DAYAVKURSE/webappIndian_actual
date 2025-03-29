package service

import (
    "BlessedApi/internal/middleware"
    "BlessedApi/pkg/logger"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
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
        c.JSON(500, gin.H{"error": "Failed to get user ID"})
        return
    }

    var input struct {
        Amount int `json:"amount"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        logger.Error("Failed to bind JSON: %v", err)
        c.JSON(400, gin.H{"error": "Invalid input format"})
        return
    }

    // Проверяем минимальную сумму
    if input.Amount < 500 {
        c.JSON(400, gin.H{"error": "Minimum amount is 500 INR"})
        return
    }

    // Создаем запрос к платежной системе
    paymentReq := PaymentRequest{
        Amount:        input.Amount,
        Currency:      "INR",
        Buttons:       []int{500, 1000, 2000, 5000},
        PaymentSystem: []string{"paytm", "phonepe", "upi_p2p"},
        CustomUserID:  fmt.Sprintf("user_%d", userID),
        ReturnURL:     fmt.Sprintf("%s?user_id=user_%d", returnURL, userID),
        Language:      "EN",
        WebhookID:     webhookID,
    }

    // Логируем запрос для отладки
    logger.Info("Sending payment request: %+v", paymentReq)

    // Конвертируем запрос в JSON
    jsonData, err := json.Marshal(paymentReq)
    if err != nil {
        logger.Error("Failed to marshal payment request: %v", err)
        c.JSON(500, gin.H{"error": "Failed to prepare payment request"})
        return
    }

    logger.Info("Request JSON: %s", string(jsonData))

    // Создаем HTTP запрос
    req, err := http.NewRequest("POST", paymentAPIURL, bytes.NewBuffer(jsonData))
    if err != nil {
        logger.Error("Failed to create request: %v", err)
        c.JSON(500, gin.H{"error": "Failed to create payment request"})
        return
    }

    req.Header.Set("Content-Type", "application/json")

    // Отправляем запрос
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    logger.Info("Sending request to: %s", paymentAPIURL)
    resp, err := client.Do(req)
    if err != nil {
        logger.Error("Failed to send request: %v", err)
        c.JSON(500, gin.H{"error": "Failed to connect to payment service"})
        return
    }
    defer resp.Body.Close()

    // Читаем тело ответа
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        logger.Error("Failed to read response body: %v", err)
        c.JSON(500, gin.H{"error": "Failed to read payment service response"})
        return
    }

    logger.Info("Response status: %d", resp.StatusCode)
    logger.Info("Response body: %s", string(body))

    if resp.StatusCode != http.StatusOK {
        logger.Error("Payment API returned non-200 status: %d", resp.StatusCode)
        c.JSON(500, gin.H{
            "error":   "Payment service error",
            "details": string(body),
        })
        return
    }

    // Парсим ответ
    var paymentResp PaymentResponse
    if err := json.Unmarshal(body, &paymentResp); err != nil {
        logger.Error("Failed to decode response: %v", err)
        c.JSON(500, gin.H{"error": "Failed to process payment service response"})
        return
    }

    if !paymentResp.Success {
        logger.Error("Payment creation failed: %+v", paymentResp)
        c.JSON(500, gin.H{
            "error":   "Payment creation failed",
            "details": paymentResp,
        })
        return
    }

    // Возвращаем URL для перенаправления
    c.JSON(200, gin.H{
        "url": paymentResp.URL,
    })
} 