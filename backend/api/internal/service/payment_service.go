package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	paymentAPIURL = "https://pay-crm.com/Remotes/create-payment-page"
	webhookID     = "abc1234"
	returnURL     = "https://blessed.one/return"
	apiKey        = "cc65c8f80dddbba8f81c5d9c3f985f07"
	accessKey     = "0801ae2d1a0e2634a0abe0e968a50f03" // Публичный ключ для вебхука
	privateKey    = "822ba2d529f8e23920b2f82944c8e870" // Приватный ключ для вебхука
)

type PaymentRequest struct {
	Amount        int      `json:"amount"`
	Currency      string   `json:"currency"`
	Buttons       []int    `json:"buttons"`
	PaymentSystem []string `json:"payment_system"`
	CustomUserID  string   `json:"custom_user_id"`
	ReturnURL     string   `json:"return_url"`
	Language      string   `json:"language"`
	WebhookID     string   `json:"webhook_id"`
}

type PaymentResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	OrderID string `json:"order_id"`
}

type WebhookRequest struct {
	AccessKey    string        `json:"access_key"`
	Signature    string        `json:"signature"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Amount              float64 `json:"amount"`
	Status              string  `json:"status"`
	Currency            string  `json:"currency"`
	OrderID             string  `json:"order_id"`
	CreatedAt           int64   `json:"created_at"`
	ActivatedAt         int64   `json:"activated_at"`
	CustomUserID        string  `json:"custom_user_id"`
	PaymentSystem       string  `json:"payment_system"`
	CustomTransactionId *string `json:"custom_transaction_id"`
}

// Проверка подписи вебхука
func verifyWebhookSignature(accessKey, signature string, transactions []Transaction) bool {
	// Конвертируем транзакции в JSON
	transactionsJSON, err := json.Marshal(transactions)
	if err != nil {
		logger.Error("Failed to marshal transactions: %v", err)
		return false
	}

	// Вычисляем MD5 от JSON транзакций
	md5Hash := md5.Sum(transactionsJSON)
	md5String := hex.EncodeToString(md5Hash[:])

	// Формируем строку для SHA1
	dataToHash := accessKey + privateKey + md5String

	// Вычисляем SHA1
	sha1Hash := sha1.Sum([]byte(dataToHash))
	calculatedSignature := hex.EncodeToString(sha1Hash[:])

	// Сравниваем с полученной подписью
	return calculatedSignature == signature
}

// Обработчик вебхука
func PaymentWebhook(c *gin.Context) {
	var webhookReq WebhookRequest
	if err := c.ShouldBindJSON(&webhookReq); err != nil {
		logger.Error("Failed to bind webhook request: %v", err)
		c.JSON(400, gin.H{"error": "Invalid webhook data"})
		return
	}

	// Проверяем подпись
	if !verifyWebhookSignature(webhookReq.AccessKey, webhookReq.Signature, webhookReq.Transactions) {
		logger.Error("Invalid webhook signature")
		c.JSON(400, gin.H{"error": "Invalid signature"})
		return
	}

	// Обрабатываем каждую транзакцию
	for _, transaction := range webhookReq.Transactions {
		if transaction.Status == "Success" {
			// Извлекаем user_id из custom_user_id
			var userID int64
			_, err := fmt.Sscanf(transaction.CustomUserID, "user_%d", &userID)
			if err != nil {
				logger.Error("Failed to parse user ID from custom_user_id: %v", err)
				continue
			}

			// TODO: Добавить логику начисления средств пользователю
			logger.Info("Processing successful transaction: %+v", transaction)
			SendMsgTg(transaction)
			// Здесь нужно добавить код для начисления средств пользователю
			// Например:
			err = addFundsToUser(userID, transaction.Amount)
			if err != nil {
				logger.Error("Failed to add funds to user: %v", err)
				continue
			}
		} else {
			logger.Info("Skipping failed/rejected transaction: %+v", transaction)
		}
	}

	// Отправляем успешный ответ
	c.JSON(200, gin.H{"status": "OK"})
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
	req.Header.Set("apikey", apiKey)

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

// Удобная обёртка, если транзакции снаружи нет
func addFundsToUser(userID int64, amount float64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		return addFundsToUserTx(tx, userID, amount)
	})
}

func addFundsToUserTx(tx *gorm.DB, userID int64, amount float64) error {
	if amount == 0 {
		return nil
	}
	if tx == nil {
		tx = db.DB
	}

	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, userID).Error; err != nil {
		return logger.WrapError(err, "failed to load user")
	}

	newBalance := user.BalanceRupee + amount

	// Обновляем только одно поле — надёжно и без гонок
	if err := tx.Model(&user).
		Update("balance_rupee", newBalance).Error; err != nil {
		return logger.WrapError(err, "failed to update user balance")
	}

	return nil
}
