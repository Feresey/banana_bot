package service

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func makeBan(msg *tgbotapi.Message) {
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
	kick := &tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: msg.Chat.ID,
			UserID: msg.ReplyToMessage.From.ID,
		}}
	kickMember(kick, reply)
}

func kickMember(kick *tgbotapi.KickChatMemberConfig, reply tgbotapi.MessageConfig) {
	ok := false
	switch kick.UserID {
	case 425496698:
		reply.Text = "Я не могу пойти против создателя. Ave Banana!"
	case 1066353768:
		reply.Text = "Бан бану рознь."
	default:
		ok = true
	}
	sendMsg(reply)

	if ok {
		resp, err := bot.KickChatMember(*kick)
		if err != nil {
			log.Warnf("%#v : [%s]", resp, err.Error())
		} else {
			log.Infof("Succeccfully kicked: ID:[%d]", kick.UserID)
		}
	}
}
