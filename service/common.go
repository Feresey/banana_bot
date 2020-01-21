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

func checkAdmins(id int64) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := AdminList[id]
	return ok
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

func getAdminsFromChat(id int64) {
	mu.Lock()
	chatAdmins, err := bot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: id})
	if err != nil {
		log.Error("Unable get chat admins", err)
	}
	AdminList[id] = chatAdmins
	mu.Unlock()
}

const (
	no = iota + 1
	yes
	undefined
)

func checkPermissions(msg *tgbotapi.Message) int {
	userID := msg.From.ID
	if userID == 425496698 || userID == 1066353768 {
		return undefined
	}
	if msg.Chat.IsPrivate() {
		return yes
	}
	chatID := msg.Chat.ID
	if !checkAdmins(chatID) {
		getAdminsFromChat(chatID)
	}
	mu.RLock()
	defer mu.RUnlock()

	for localID, admins := range AdminList {
		if localID != chatID {
			continue
		}
		for _, admin := range admins {
			if admin.User.ID == userID {
				return yes
			}
		}
	}
	return no
}

func joke(userID int) string {
	switch userID {
	case 425496698:
		return "Я не могу пойти против создателя. Ave Banana!"
	case 1066353768:
		return "Бан бану рознь."
	default:
		return ""
	}
}
