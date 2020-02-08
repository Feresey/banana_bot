package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

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

func kickMember(p *model.Person) error {
	kick := &tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: p.ChatID,
			UserID: p.UserID,
		},
	}

	resp, err := bot.KickChatMember(*kick)
	if err != nil {
		log.Warnf("%#v : [%s]", resp, err.Error())
	} else {
		log.Infof("Succeccfully kicked: ID:[%d]", kick.UserID)
	}
	return nil
}
