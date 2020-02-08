package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Feresey/banana_bot/db"
	"github.com/Feresey/banana_bot/model"
)

func sendMsg(msg tgbotapi.Chattable) {
	resp, err := bot.Send(msg)
	if err != nil {
		log.Errorf("Unable to send message: %s\n %#v", err, resp)
	}
}

func isPublicMethod(cmd string) bool {
	switch cmd {
	case "report", "subscribe", "unsubscribe", "":
		return true
	default:
		return false
	}
}

func isAdmin(msg model.Message) bool {
	if msg.Chat.IsPrivate() {
		return true
	}

	member, err := bot.GetChatMember(tgbotapi.ChatConfigWithUser{ChatID: msg.Chat.ID, UserID: msg.From.ID})
	if err != nil {
		log.Error("Unable get info about user", err)
	}
	return member.IsAdministrator() || member.IsCreator()
}

func addNewChatIfNeeded(chatid int64) {
	list, err := db.GetChatList()
	if err != nil {
		log.Error(err)
		return
	}

	for _, val := range list {
		if val == chatid {
			return
		}
	}

	updateAdminsFromChat(chatid)
}

func updateAdminsFromChat(chatid int64) []int {
	members, err := bot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: chatid})
	if err != nil {
		log.Warn("Unable update admins", err)
		return nil
	}

	ids := make([]int, 0, len(members))
	for idx := range members {
		ids = append(ids, members[idx].User.ID)
	}

	err = db.SetAdmins(chatid, ids)
	if err != nil {
		log.Error("Unable upate admins", err)
	}
	return ids
}

func updateAllAdmins() {
	chats, err := db.GetChatList()
	if err != nil {
		log.Error(err)
		return
	}

	for _, val := range chats {
		updateAdminsFromChat(val)
	}
}
