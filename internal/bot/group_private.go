package bot

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Feresey/banana_bot/internal/db"
	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	day     = 24 * time.Hour
	forever = 0
)

const (
	godID = 425496698
)

type ErrProtected struct {
	Who string
}

func (e ErrProtected) Error() string { return fmt.Sprintf("protected call: %s", e.Who) }

var (
	protectGodMessage  = "Я не могу пойти против создателя. Ave Banana!"
	protectMeMessage   = "Бан бану рознь."
	protectSelfMessage = "{{formatUser .}}, мазохизм не приветствуется."

	needReplyMessage = "Надо использовать команду ответом на сообщение"
)

func (b *Bot) protect(target *db.Person, callMessage *tgbotapi.Message) error {
	var reply NeedFormat

	ok := false

	switch target.UserID {
	case godID:
		reply.Message = protectGodMessage
	case int64(b.me.ID):
		reply.Message = protectMeMessage
	case int64(callMessage.From.ID):
		reply.Message = protectSelfMessage
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
		return ErrProtected{Who: callMessage.From.String()}
	}
	return nil
}

func (b *Bot) needReply(msg *tgbotapi.Message) error {
	return b.ReplyOne(msg, NeedFormat{Message: needReplyMessage})
}

func getKickConfig(p *db.Person, until time.Duration) tgbotapi.KickChatMemberConfig {
	return tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: p.ChatID,
			UserID: int(p.UserID),
		},
		UntilDate: time.Now().Add(until).Unix(),
	}
}

func (b *Bot) kickMember(p *db.Person, until time.Duration) error {
	_, err := b.api.KickChatMember(getKickConfig(p, until))
	b.log.Info("Kick person",
		zap.Int64("user_id", p.UserID), zap.Int64("chat_id", p.ChatID),
		zap.Duration("until", until),
		zap.Error(err))
	return err
}

func formatKick(target *tgbotapi.User, until time.Duration) NeedFormat {
	toMsg := "навеки.\nF"
	if until != forever {
		toMsg = fmt.Sprintf("на %s", until.String())
	}
	return NeedFormat{
		Message: "{{formatUser .User}} забанен {{.To}}.",
		FormatParams: map[string]interface{}{
			"User": target,
			"To":   toMsg,
		},
	}
}

func (b *Bot) kick(ctx context.Context, msg *tgbotapi.Message, until time.Duration) error {
	if msg.ReplyToMessage == nil {
		return b.needReply(msg)
	}

	targetUser := msg.ReplyToMessage.From
	target := &db.Person{
		ChatID: msg.Chat.ID,
		UserID: int64(targetUser.ID),
	}

	if err := b.protect(target, msg); err != nil {
		b.log.Warn("Protect", zap.Error(err))
		return nil
	}

	_, err := b.api.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.MessageID})
	if err != nil {
		b.log.Error("Delete warns message", zap.Error(err))
	}

	if err := b.kickMember(target, until); err != nil {
		return err
	}

	return b.ToChat(msg.Chat.ID, formatKick(targetUser, until))
}

func (b *Bot) limitation(target *db.Person, total int64) tgbotapi.RestrictChatMemberConfig {
	f := false
	until := big.NewInt(b.c.MaxWarn)
	until.Exp(until, big.NewInt(total), nil)

	return tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: target.ChatID,
			UserID: int(target.UserID),
		},
		CanSendMediaMessages:  &f,
		CanAddWebPagePreviews: &f,
		CanSendOtherMessages:  &f,
		CanSendMessages:       &f,

		UntilDate: time.Now().Add(time.Minute * time.Duration(until.Int64())).Unix(),
	}
}

func formatWarn(target *tgbotapi.User, total int64, max int64) NeedFormat {
	return NeedFormat{
		Message: "{{formatUser .User}}, Предупреждение {{.Total}}/{{.MaxWarn}}",
		FormatParams: map[string]interface{}{
			"User":    target,
			"Total":   total,
			"MaxWarn": max,
		},
	}
}

func formatUnwarn(target *tgbotapi.User, total int64) NeedFormat {
	return NeedFormat{
		Message: "{{formatUser .User}} вёл себя хорошо и у него сняли одно предупреждение. " +
			"Всего предупреждений: {{.Total}}",
		FormatParams: map[string]interface{}{
			"User":  target,
			"Total": total,
		},
	}
}

func formatLastWarn(target *tgbotapi.User) NeedFormat {
	return NeedFormat{
		Message:      "{{formatUser .}}, Последнее предупреждение",
		FormatParams: target,
	}
}

func (b *Bot) warn(ctx context.Context, msg *tgbotapi.Message, add bool) error {
	log := b.log.With(zap.Stringer("caller", msg.From), zap.Int("message_id", msg.MessageID))
	log.Info("Warn")
	if msg.ReplyToMessage == nil {
		return b.needReply(msg)
	}
	log.Info("Caller", zap.Stringer("target", msg.ReplyToMessage.From))

	targetUser := msg.ReplyToMessage.From
	target := &db.Person{
		ChatID: msg.Chat.ID,
		UserID: int64(targetUser.ID),
	}

	if err := b.protect(target, msg); err != nil {
		b.log.Warn("Protect", zap.Error(err))
		return nil
	}

	log.Debug("Database")
	total, err := b.db.Warn(ctx, target, add)
	if err != nil {
		return err
	}

	log.Debug("Response")
	// плюшки
	if !add {
		return b.ToChat(msg.Chat.ID, formatUnwarn(targetUser, total))
	}
	if _, err := b.api.RestrictChatMember(b.limitation(target, total)); err != nil {
		log.Error("Restrict", zap.Error(err))
	}

	_, err = b.api.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.MessageID})
	if err != nil {
		b.log.Error("Delete warns message", zap.Error(err))
	}

	var reply NeedFormat

	switch {
	case total < b.c.MaxWarn:
		reply = formatWarn(targetUser, total, b.c.MaxWarn)
	case total == b.c.MaxWarn:
		reply = formatLastWarn(targetUser)
	case total > b.c.MaxWarn:
		reply = formatKick(targetUser, forever)
		if err := b.kickMember(target, forever); err != nil {
			b.log.Error("Kick person", zap.Error(err))
		}
	}
	return b.ToChat(msg.Chat.ID, reply)
}
