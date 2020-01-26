package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Feresey/banana_bot/db"
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
	case "report", "":
		return true
	default:
		return false
	}
}

func (b *Bot) isAdmin(msg *model.Message) bool {
	if msg.Chat.IsPrivate() {
		return true
	}

	member, err := b.GetChatMember(tgbotapi.ChatConfigWithUser{ChatID: msg.Chat.ID, UserID: msg.From.ID})
	if err != nil {
		b.log.Error("Unable get info about user", err)
	}
	return member.IsAdministrator() || member.IsCreator()
}

func (b *Bot) updateAdminsFromChat(chatid int64) []int {
	members, err := b.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: chatid})
	if err != nil {
		b.log.Warn("Unable update admins", err)
		return nil
	}

	ids := make([]int, 0, len(members))
	for idx := range members {
		ids = append(ids, members[idx].User.ID)
	}

	err = db.SetAdmins(chatid, ids)
	if err != nil {
		b.log.Error("Unable upate admins", err)
	}
	return ids
}

func (b *Bot) updateAllAdmins() {
	chats, err := db.GetChatList()
	if err != nil {
		b.log.Error(err)
		return
	}

	for _, val := range chats {
		b.updateAdminsFromChat(val)
	}
}
