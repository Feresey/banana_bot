package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Feresey/banana_bot/db"
	"github.com/Feresey/banana_bot/format"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	day     = 24 * time.Hour
	forever = 0
)

var ErrProtected = errors.New("protected call")

func (b *Bot) needReply(msg tgbotapi.Message) error {
	return b.ReplyOne(msg, format.NeedFormat{
		Message: "Надо использовать команду ответом на сообщение"})
}

func (b *Bot) kickMember(p *db.Person, kickTime time.Duration) error {
	kick := &tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: p.ChatID,
			UserID: int(p.UserID),
		},
		UntilDate: time.Now().Add(kickTime).Unix(),
	}
	_, err := b.api.KickChatMember(*kick)
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

	return b.ToChat(msg.Chat.ID, format.NeedFormat{
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

	var reply format.NeedFormat

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
