package service

import (
	"fmt"

	"github.com/Feresey/bot-tg/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func makeWarn(msg *tgbotapi.Message, add bool) {
	if msg.ReplyToMessage == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Не указано кому /warn кидать")
		sendMsg(reply)
		return
	}

	total, err := db.Warn(msg.ReplyToMessage.From.ID, add)
	if err != nil {
		log.Warn(err)
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("@%s /warn [%d/5]", msg.ReplyToMessage.From.UserName, total))
	sendMsg(reply)
}
