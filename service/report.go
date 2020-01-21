package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func makeReportAdmins(msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "@Feresey")
	sendMsg(reply)
}
