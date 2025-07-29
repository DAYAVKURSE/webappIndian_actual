package service

import (
	"BlessedApi/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
	"time"
)

var Tg *tgbotapi.BotAPI
var ChatId int64

func InitTg() {
	botToken := os.Getenv("BOT_TOKEN")
	chatID, _ := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Error("telegram error: %v", err)
	}
	Tg = bot
	ChatId = chatID
}

func SendMsgTg(transaction Transaction) {
	msg := tgbotapi.NewMessage(ChatId, makeMessage(transaction))
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	if _, err := Tg.Send(msg); err != nil {
		log.Printf("telegram send: %v", err)
	}
}

func makeMessage(d Transaction) string {
	tm := time.Unix(d.CreatedAt, 0).Format("02.01.2006 15:04:05")
	return "💸 *Поступление средств*\n" +
		"`" + d.OrderID + "`\n" +
		"Сумма: *" + strconv.FormatFloat(d.Amount, 'f', 2, 64) + " " + d.Currency + "*\n" +
		"Дата: " + tm
}
