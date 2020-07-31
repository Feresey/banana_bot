package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"

	"github.com/Feresey/banana_bot/internal/db"
	"github.com/Feresey/banana_bot/internal/format"
)

func personFromMessage(msg tgbotapi.Message) *db.Person {
	return &db.Person{
		ChatID: msg.Chat.ID,
		UserID: int64(msg.From.ID),
	}
}

func (b *Bot) ReplyOne(forward tgbotapi.Message, need format.NeedFormat) error {
	return format.New(
		b.api, tgbotapi.BaseChat{ChatID: forward.Chat.ID, ReplyToMessageID: forward.MessageID},
	).Format(need)
}

func (b *Bot) ToChat(chatID int64, need format.NeedFormat) error {
	return format.New(
		b.api, tgbotapi.BaseChat{ChatID: chatID},
	).Format(need)
}

func (b *Bot) isPublicMethod(cmd string) bool {
	switch cmd {
	case "report", "subscribe", "unsubscribe":
		return true
	default:
		return false
	}
}

func (b *Bot) isAdmin(msg tgbotapi.Message) bool {
	if msg.Chat.IsPrivate() {
		return true
	}
	member, err := b.api.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: msg.Chat.ID,
		UserID: msg.From.ID,
	})
	if err != nil {
		b.log.Error("Unable get info about user", zap.Error(err))
	}
	// safe to call on default value
	return member.IsAdministrator() || member.IsCreator()
}

const (
	godID = 425496698
	myID  = 1066353768
)

func (b *Bot) protect(target *db.Person, callMessage tgbotapi.Message) error {
	var reply format.NeedFormat

	ok := false

	switch target.UserID {
	case godID:
		reply.Message = "Я не могу пойти против создателя. Ave Banana!"
	case myID:
		reply.Message = "Бан бану рознь."
	case int64(callMessage.From.ID):
		reply.Message = "{{formatUser .}}, мазохизм не приветствуется."
		reply.FormatParams = callMessage.From
	default:
		ok = true
	}

	if reply.Message != "" {
		if err := b.ReplyOne(callMessage, reply); err != nil {
			return err
		}
	}

	if !ok {
		return fmt.Errorf("%w: %s", ErrProtected, callMessage.From.UserName)
	}
	return nil
}
