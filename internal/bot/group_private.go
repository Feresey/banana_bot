package bot

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Feresey/banana_bot/internal/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	day     = 24 * time.Hour
	forever = 0
)

const (
	godID = 425496698
	myID  = 1066353768
)

var ErrProtected = errors.New("protected call")

func (b *Bot) protect(target *db.Person, callMessage tgbotapi.Message) error {
	var reply NeedFormat

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

func (b *Bot) needReply(msg tgbotapi.Message) error {
	return b.ReplyOne(msg, NeedFormat{
		Message: "Надо использовать команду ответом на сообщение"})
}

func (b *Bot) kickMember(p *db.Person, kickTime time.Duration) error {
	kick := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: p.ChatID,
			UserID: int(p.UserID),
		},
		UntilDate: time.Now().Add(kickTime).Unix(),
	}
	_, err := b.api.KickChatMember(kick)
	return err
}

func (b *Bot) kick(ctx context.Context, msg tgbotapi.Message, expire time.Duration) error {
	if msg.ReplyToMessage == nil {
		return b.needReply(msg)
	}

	targetUser := msg.ReplyToMessage.From
	target := &db.Person{
		ChatID: msg.Chat.ID,
		UserID: int64(targetUser.ID),
	}

	if err := b.protect(target, msg); err != nil {
		return err
	}
	if err := b.kickMember(target, expire); err != nil {
		return err
	}

	toMsg := "навеки"
	if expire != forever {
		toMsg = fmt.Sprintf("на %s", expire.String())
	}

	return b.ToChat(msg.Chat.ID, NeedFormat{
		Message: "{{formatUser .User}} забанен {{.To}}.\nF",
		FormatParams: map[string]interface{}{
			"User": targetUser,
			"To":   toMsg,
		},
	})
}

func (b *Bot) warn(ctx context.Context, msg tgbotapi.Message, add bool) error {
	if msg.ReplyToMessage == nil {
		return b.needReply(msg)
	}

	targetUser := msg.ReplyToMessage.From
	target := &db.Person{
		ChatID: msg.Chat.ID,
		UserID: int64(targetUser.ID),
	}

	if err := b.protect(target, msg); err != nil {
		return err
	}

	total, err := b.db.Warn(ctx, target, add)
	if err != nil {
		return err
	}

	if add {
		f := false
		until := big.NewInt(b.c.MaxWarn)
		until.Exp(until, big.NewInt(total), nil)

		conf := tgbotapi.RestrictChatMemberConfig{}
		conf.UserID = int(target.UserID)
		conf.ChatID = target.ChatID
		conf.UntilDate = time.Now().Add(time.Minute * time.Duration(until.Int64())).Unix()
		conf.CanSendMessages = &f
		conf.CanSendMediaMessages = &f
		conf.CanSendOtherMessages = &f
		conf.CanAddWebPagePreviews = &f

		if _, err := b.api.RestrictChatMember(conf); err != nil {
			return err
		}
	}

	var reply NeedFormat

	switch {
	case total < b.c.MaxWarn:
		reply.Message = "{{formatUser .User}}, Предупреждение {{.Total}}/{{.MaxWarn}}"
		reply.FormatParams = map[string]interface{}{
			"User":    targetUser,
			"Total":   total,
			"MaxWarn": b.c.MaxWarn,
		}
	case total == b.c.MaxWarn:
		reply.Message = "{{formatUser .}}, Последнее предупреждение"
		reply.FormatParams = targetUser
	case total > b.c.MaxWarn:
		return b.kickMember(target, forever)
	}
	return b.ToChat(msg.Chat.ID, reply)
}
