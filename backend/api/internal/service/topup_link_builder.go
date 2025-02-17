package service

import (
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	PaymentAPIURL = "https://pay-crm.com/Remotes/create-payment-page"
)

var (
	APIKey     string
	FrontendIP string
	ok         bool
)

func init() {
	APIKey, ok = os.LookupEnv("APIKEY")
	if !ok {
		logger.Error("APIKEY environment variable not set")
	}

	FrontendIP, ok = os.LookupEnv("FRONTEND_IP")
	if !ok {
		logger.Error("FRONTEND_IP environment variable not set")
	}
}

type PaymentPageRequest struct {
	Amount       int    `json:"amount"`
	Buttons      []int  `json:"buttons"`
	Currency     string `json:"currency"`
	CustomUserID string `json:"custom_user_id"`
	ReturnURL    string `json:"return_url"`
}

type AmountRequest struct {
	Amount int `json:"amount"`
}

type PaymentPageResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	OrderID string `json:"order_id"`
}

func CreatePaymentPageHandler(c *gin.Context) {
	// Get user ID using your middleware
	userID, err := middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Check if user exists
	var exists bool
	exists, err = models.CheckIfUserExistsByID(userID)
	if err != nil {
		logger.Error("unable to check if user with id %d exists: %v", userID, err)
		c.Status(500)
		return
	}

	if !exists {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	var amountReq AmountRequest
	if err := c.ShouldBindJSON(&amountReq); err != nil {
		logger.Error("invalid request body: %v", err)
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if amountReq.Amount < models.MinDepositRupee {
		c.JSON(406, gin.H{"error": "Minumim deposit is 500 rupees"})
		return
	}

	// Create payment page request
	paymentReq := PaymentPageRequest{
		Amount:       amountReq.Amount,
		Buttons:      []int{300, 500, 1000},
		Currency:     "INR",
		CustomUserID: fmt.Sprintf("%d", userID), // Convert uint to string
		ReturnURL:    FrontendIP + "/wallet",
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(paymentReq)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	logger.Debug("Payment request: %s", string(jsonData))

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("POST", PaymentAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Apikey", APIKey)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Debug("payment API returned status: %d", resp.StatusCode)
		c.Status(resp.StatusCode)
		return
	}

	// Decode response
	var paymentResp PaymentPageResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	logger.Debug("Payment Info: %v", paymentResp)

	// Return the payment URL to frontend
	c.JSON(200, gin.H{"url": paymentResp.URL})
}
