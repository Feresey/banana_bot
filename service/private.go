package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func startPrivateChat(msg *tgbotapi.Message) {
	response := tgbotapi.NewMessage(msg.Chat.ID, "Приветствую, кожаный мешок.")
	_, err := bot.Send(response)
	if err != nil {
		log.Error("Unable to send message:", err)
	}
}
