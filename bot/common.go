package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Feresey/banana_bot/model"
)

func (b *Bot) sendMsg(msg tgbotapi.Chattable) {
	_, err := b.Send(msg)
	if err != nil {
		b.log.Error("Unable to send message:", err)
	}
}

func (b *Bot) isPublicMethod(cmd string) bool {
	switch cmd {
	case "report":
		return true
	default:
		return false
	}
}

func (b *Bot) getAdmins(id int64) ([]tgbotapi.ChatMember, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	members, ok := b.adminList[id]

	return members, ok
}

func (b *Bot) updateAdmins() {
	b.mu.RLock()
	admins := []int64{}
	for chatID := range b.adminList {
		admins = append(admins, chatID)
	}
	b.mu.RUnlock()

	for _, chatID := range admins {
		b.updateAdminsFromChat(chatID)
	}
}

func (b *Bot) updateAdminsFromChat(id int64) []tgbotapi.ChatMember {
	b.mu.Lock()
	defer b.mu.Unlock()

	chatAdmins, err := b.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: id})
	if err != nil {
		b.log.Error("Unable get chat admins", err)
	}
	b.adminList[id] = chatAdmins
	return chatAdmins
}

func (b *Bot) isAdmin(msg *model.Message) bool {
	if msg.Chat.IsPrivate() {
		return true
	}

	chatID := msg.Chat.ID
	admins, ok := b.getAdmins(chatID)
	if !ok {
		admins = b.updateAdminsFromChat(chatID)
	}

	userID := msg.From.ID
	for _, pipl := range admins {
		if pipl.User.ID == userID {
			return true
		}
	}
	return false
}

// func joke(userID int) string {
// 	switch userID {
// 	case 425496698:
// 		return "Я не могу пойти против создателя. Ave Banana!"
// 	case 1066353768:
// 		return "Бан бану рознь."
// 	default:
// 		return ""
// 	}
// }
