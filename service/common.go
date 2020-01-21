package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func sendMsg(msg tgbotapi.MessageConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Error("Unable to send message:", err)
	}
}

func getAdmins() {
	mu.Lock()
	for chat := range AdminList {
		chatAdmins, err := bot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: chat})
		if err != nil {
			log.Error("Unable get chat admins", err)
			continue
		}
		AdminList[chat] = chatAdmins
	}
	mu.Unlock()
}

func isAdmin(msg *tgbotapi.Message) bool {
	mu.RLock()
	defer mu.RUnlock()

	for chatID, admins := range AdminList {
		if chatID != msg.Chat.ID {
			continue
		}
		for _, admin := range admins {
			if admin.User.ID == msg.From.ID {
				return true
			}
		}
	}
	return false
}
