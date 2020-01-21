package service

import (
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func makeBan(msg *tgbotapi.Message) {
	switch checkPermissions(msg) {
	case yes:
		appeal := ""
		if msg.Entities != nil {
			for _, entity := range *msg.Entities {
				if entity.Type == "mention" && len(msg.Text) <= entity.Offset+entity.Length {
					appeal = msg.Text[entity.Offset : entity.Offset+entity.Length]
					break
				}
			}
		}

		reply := tgbotapi.NewMessage(msg.Chat.ID, "")

		if msg.ReplyToMessage == nil {
			if appeal != "" {
				reply.Text = fmt.Sprintf("Очень хочу забанить %s, но надо вызывать команду ответом на сообщение", appeal)
			} else {
				reply.Text = "Команду нужно использовать ответом на сообщение"
				reply.ReplyToMessageID = msg.MessageID
			}
			sendMsg(reply)
			return
		}

		reply.Text = fmt.Sprintf("Пора забанить @%s", msg.ReplyToMessage.From.UserName)

		err := kickMember(msg.Chat.ID, msg.ReplyToMessage.From.ID)
		if err != nil {
			reply.Text = err.Error()
		}
		sendMsg(reply)
	case no:
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Только админам можно")
		reply.ReplyToMessageID = msg.MessageID
		sendMsg(reply)
	case undefined:
		reply := tgbotapi.NewMessage(msg.Chat.ID, joke(msg.From.ID))
		reply.ReplyToMessageID = msg.MessageID
		sendMsg(reply)
	}
}

func kickMember(chatID int64, userID int) error {
	kick := &tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		}}

	if s := joke(userID); s != "" {
		return errors.New(s)
	}

	resp, err := bot.KickChatMember(*kick)
	if err != nil {
		log.Warnf("%#v : [%s]", resp, err.Error())
	} else {
		log.Infof("Succeccfully kicked: ID:[%d]", kick.UserID)
	}
	return nil
}
