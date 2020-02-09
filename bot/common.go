package bot

import (
	"time"

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
	case "report", "subscribe", "unsubscribe":
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

const (
	day     = 24 * time.Hour
	forever = 0
)

func kickMember(p *model.Person, kickTime time.Duration) error {
	kick := &tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: p.ChatID,
			UserID: p.UserID,
		},
		UntilDate: time.Now().Add(kickTime).Unix(),
	}

	resp, err := bot.KickChatMember(*kick)
	if err != nil {
		log.Warnf("%#v : [%s]", resp, err.Error())
	} else {
		log.Infof("Succeccfully kicked: %#v", kick)
	}
	return nil
}
