package service

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"BlessedApi/internal/models"
	"BlessedApi/pkg/logger"

	"github.com/gin-gonic/gin"
)

type TransactionBody struct {
	Amount              float64 `json:"amount"`
	Status              string  `json:"status"`
	Currency            string  `json:"currency"`
	OrderID             string  `json:"order_id"`
	CreatedAt           int64   `json:"created_at"`
	ActivatedAt         *int64  `json:"activated_at"`
	CustomUserID        *string `json:"custom_user_id"`
	PaymentSystem       string  `json:"payment_system"`
	CustomTransactionID *string `json:"custom_transaction_id"`
}

type PostbackBody struct {
	AccessKey    string            `json:"access_key"`
	Signature    string            `json:"signature"`
	Transactions []TransactionBody `json:"transactions"`
}

type NotificationBody struct {
	UserID string  `json:"userId"`
	Amount float64 `json:"amount"`
}

func sendNotification(userID string, amount float64) error {
	notification := NotificationBody{
		UserID: userID,
		Amount: amount,
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %v", err)
	}

	resp, err := http.Post("http://depositBot:8787/notification", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification server returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func PaymentSystemPostback(c *gin.Context) {
	var postbackBody PostbackBody

	if err := c.ShouldBindJSON(&postbackBody); err != nil {
		body, _ := io.ReadAll(c.Request.Body)
		logger.Error("Unable to unmarshal postback: %s", string(body))
		c.JSON(500, gin.H{"error": "Unable to unmarshal body"})
		return
	}

	body, _ := io.ReadAll(c.Request.Body)
	logger.Debug("Postback: %s", string(body))

	postbackBodyStr, _ := json.MarshalIndent(postbackBody, "", "\t")
	logger.Debug("%s", postbackBodyStr)

	// First verify if access key exists and is valid
	expectedAccessKey := os.Getenv("ACCESS_KEY")
	if postbackBody.AccessKey != expectedAccessKey {
		c.JSON(403, gin.H{"error": "invalid access key"})
		return
	}

	signatureVerified, err := verifySignature(
		postbackBody.AccessKey,
		postbackBody.Transactions,
		postbackBody.Signature)
	if err != nil {
		logger.Debug("Unable to verify signature: %v", err)
		c.JSON(401, gin.H{"error": "unable to verify signature"})
		return
	}

	if !signatureVerified {
		c.JSON(403, gin.H{"error": "signature not valid"})
		return
	}

	successfullTransactions := 0

	for i := range postbackBody.Transactions {
		isWithdrawal, err := models.UpdateWithdrawalStatusIfRequired(
			postbackBody.Transactions[i].OrderID,
			postbackBody.Transactions[i].Status)
		if err != nil {
			logger.Error("%v", err)
			continue
		}

		if isWithdrawal {
			successfullTransactions++
		}

		if !isWithdrawal && postbackBody.Transactions[i].Status == "Success" {
			if AddDeposit(c, postbackBody.Transactions[i]) {
				successfullTransactions++

				// Send notification for successful deposit
				if postbackBody.Transactions[i].CustomUserID != nil {
					err := sendNotification(*postbackBody.Transactions[i].CustomUserID, postbackBody.Transactions[i].Amount)
					if err != nil {
						logger.Error("Failed to send notification: %v", err)
					} else {
						logger.Debug("Sent notification for user %s with amount %f", *postbackBody.Transactions[i].CustomUserID, postbackBody.Transactions[i].Amount)
					}
				}
			}
		} else if !isWithdrawal {
			logger.Error("postback transaction status not 'Success'; OrderID: %s", postbackBody.Transactions[i].OrderID)
			// Transaction is successful even if status is Failed to give right response
			// to the payment system
			successfullTransactions++
		}
	}

	if successfullTransactions == len(postbackBody.Transactions) {
		c.JSON(200, gin.H{"status": "OK"})
	}
}

func verifySignature(
	accessKey string, transactions []TransactionBody,
	incomingSignature string) (bool, error) {
	// Get private key from environment
	privateKey, ok := os.LookupEnv("PRIVATE_KEY")
	if !ok {
		return false, logger.WrapError(errors.New("unable to get private key from env"), "")
	}

	// Marshal only the transactions array, using the same flags as the client
	transactionsJSON, err := json.Marshal(transactions)
	if err != nil {
		return false, fmt.Errorf("error marshaling transactions: %v", err)
	}

	// Calculate MD5 of the transactions JSON
	md5Hash := md5.New()
	md5Hash.Write(transactionsJSON)
	transactionsMD5 := hex.EncodeToString(md5Hash.Sum(nil))

	// Concatenate in the correct order: access_key + private_key + md5(transactions_json)
	dataToHash := accessKey + privateKey + transactionsMD5

	// Calculate SHA1
	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(dataToHash))
	calculatedSignature := hex.EncodeToString(sha1Hash.Sum(nil))
	logger.Debug("Calculated signature: %s", calculatedSignature)

	// Compare signatures
	return calculatedSignature == incomingSignature, nil
}
