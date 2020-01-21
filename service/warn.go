package service

import (
	"fmt"

	"github.com/Feresey/banana_bot/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func makeWarn(msg *tgbotapi.Message, add bool) {
	if !isAdminOrPrivate(msg) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Только админам можно")
		reply.ReplyToMessageID = msg.MessageID
		sendMsg(reply)
		return
	}
	if msg.ReplyToMessage == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Не указано кому /warn кидать")
		sendMsg(reply)
		return
	}

	total, err := db.Warn(msg.ReplyToMessage.From.ID, add)
	if err != nil {
		log.Warn(err)
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	user := msg.ReplyToMessage.From
	switch {
	case total < 5:
		reply.Text = fmt.Sprintf("@%s Предупреждение %d/5", user.UserName, total)
	case total == 5:
		reply.Text = fmt.Sprintf("@%s Последнее предупреждение!", user.UserName)
	default:
		reply.Text = "F"
	}

	if total > 5 {
		err = kickMember(msg.Chat.ID, user.ID)
		if err != nil {
			reply.Text = err.Error()
		}
	}
	sendMsg(reply)
}
