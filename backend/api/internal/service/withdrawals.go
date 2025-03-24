package service

import (
	"BlessedApi/cmd/db"
	"BlessedApi/internal/middleware"
	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var withdrawalAllowedPaymentSystems = map[string]bool{
	"imps": true,
	"neft": true,
	"rtgs": true,
	"upi": true,
}

const (
	WithdrawalAPIURL = "https://api.a-pay.one/Remotes/create-withdrawal"
)

type withdrawalInput struct {
	Amount        float64     `json:"amount" validate:"required"`
	Currency      string      `json:"currency"`
	PaymentSystem string      `json:"payment_system" validate:"required"`
	CustomUserID  string      `json:"custom_user_id"`
	Data          interface{} `json:"data"`
}

type DepositRequirementError struct {
	Required float64
	Current  float64
}

func (e *DepositRequirementError) Error() string {
	return fmt.Sprintf("You need to deposit account firstly for: 13.000 Rupees")
}

type withdrawalResponse struct {
	Success       bool    `json:"success"`
	Status        string  `json:"status"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentSystem string  `json:"payment_system"`
	CustomUserID  string  `json:"custom_user_id"`
	CreatedAt     int64   `json:"created_at"`
}

func (i *withdrawalInput) Validate() error {
	validate = validator.New()
	return validate.Struct(i)
}

func CreateWithdrawal(c *gin.Context) {
	var err error
	var input withdrawalInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Unable to unmarshal body"})
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if _, ok := withdrawalAllowedPaymentSystems[input.PaymentSystem]; !ok {
		c.JSON(400, gin.H{"error": "payment system not supported"})
		return
	}

	var user models.User
	user.ID, err = middleware.GetUserIDFromGinContext(c)
	if err != nil {
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	input.Currency = "INR"
	input.CustomUserID = strconv.FormatInt(user.ID, 10)

	errInsufficientBalance := errors.New("insufficient balance")

	//errDepositRequirementNotMet := errors.New("minimum total deposit requirement not met")
	var withdrawal models.Withdrawal

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		totalUserDep, err := models.GetUserTotalDeposit(tx, withdrawal.UserID)
		if err != nil {
			return logger.WrapError(err, "")
		}

		if totalUserDep < models.MinDepositSumToWithdrawal {
			return &DepositRequirementError{
				Required: models.MinDepositSumToWithdrawal,
				Current:  totalUserDep,
			}
		}
		if totalUserDep >= models.MinDepositSumToWithdrawal {
			return &DepositRequirementError{
				Required: models.MinDepositSumToWithdrawal,
				Current:  totalUserDep,
			}
		}
		if err = tx.First(&user, user.ID).Error; err != nil {
			return logger.WrapError(err, "")
		}

		if user.BalanceRupee < input.Amount {
			return errInsufficientBalance
		}

		user.BalanceRupee -= input.Amount
		if err = tx.Save(&user).Error; err != nil {
			return logger.WrapError(err, "")
		}

		withdrawal = models.Withdrawal{
			UserID:    user.ID,
			Amount:    input.Amount,
			Status:    "Pending",
			CreatedAt: time.Now().Unix(),
		}

		if err = tx.Create(&withdrawal).Error; err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	})

	if err != nil {
		if depErr, ok := err.(*DepositRequirementError); ok {
			c.JSON(402, gin.H{"error": depErr.Error()})
			return
		} else if errors.Is(err, errInsufficientBalance) {
			c.JSON(402, gin.H{"error": err.Error()})
			return
		} else {
			logger.Error("%v", err)
			c.Status(500)
			return
		}
	}

	sendWithdrawalRequest(c, &input, &withdrawal)
}

func sendWithdrawalRequest(c *gin.Context, input *withdrawalInput, withdrawal *models.Withdrawal) {
	var withdrawalResp withdrawalResponse
	errBadRespStatus := errors.New("response status is rejected or failed")
	errStatusBadRequest := errors.New("response status code not 200")

	err := func() error {
		jsonData, err := json.Marshal(input)
		if err != nil {
			return logger.WrapError(err, "")
		}

		logger.Debug("withdrawal input: %s", string(jsonData))

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, err := http.NewRequest("POST", WithdrawalAPIURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return logger.WrapError(err, "")
		}

		projectID, ok := os.LookupEnv("PROJECT_ID")
		if !ok {
			return errors.New("unable to get environment variable PROJECT_ID")
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("apikey", APIKey)

		q := req.URL.Query()
		q.Add("project_id", projectID)
		req.URL.RawQuery = q.Encode()

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			return logger.WrapError(err, "")
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			var jsonBody map[string]interface{}
			_ = json.Unmarshal(body, &jsonBody)

			c.JSON(resp.StatusCode, jsonBody)
			return errStatusBadRequest
		}

		// Decode response
		if err = json.NewDecoder(resp.Body).Decode(&withdrawalResp); err != nil {
			return logger.WrapError(err, "")
		}

		logger.Debug("Withdrawal Info: %v", withdrawalResp)

		if withdrawalResp.Status == "Rejected" || withdrawalResp.Status == "Failed" {
			c.JSON(202, gin.H{"status": withdrawalResp.Status})
			return errBadRespStatus
		}

		withdrawal.OrderID = withdrawalResp.OrderID
		withdrawal.Status = withdrawalResp.Status

		if err = db.DB.Save(withdrawal).Error; err != nil {
			return logger.WrapError(err, "")
		}

		return nil
	}()

	if err != nil {
		if errors.Is(err, errBadRespStatus) {
			return
		}
		logger.Error("%v", err)
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{"status": "withdrawal request created successfully"})
}
