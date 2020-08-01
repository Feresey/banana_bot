package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

var (
	ErrWTF           = errors.New("what? not group, not private. what is it???")
	ErrNoSuchCommand = errors.New("no such command")
)

var (
	onlyAdminsMessage = "Только админам можно"
)

func (b *Bot) processMessage(msg tgbotapi.Message) error {
	// предполагая что у меня руки из жопы я оставлю это здесь
	defer func() {
		if err := recover(); err != nil {
			b.log.Error("Fall in panic", zap.Any("panic", err))
		}
	}()

	b.log.Debug("Message from",
		zap.String("chat", msg.Chat.Title),
		zap.String("username", msg.From.UserName),
		zap.String("text", msg.Text),
	)

	switch chat := msg.Chat; {
	case chat.IsPrivate():
		return b.privateMessage(msg)
	case chat.IsGroup() || chat.IsSuperGroup():
		del, err := b.groupMessage(msg)
		if err != nil {
			return err
		}
		if !del {
			return nil
		}

		_, err = b.api.DeleteMessage(tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID))
		return err
	default:
		return ErrWTF
	}
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
		return false
	}
	// safe to call on default value
	return member.IsAdministrator() || member.IsCreator()
}

func (b *Bot) groupMessage(msg tgbotapi.Message) (del bool, err error) {
	if !msg.IsCommand() {
		// Обычное сообщение, лог ненужон
		return false, nil
	}
	// сообщение другому боту, или мисклик
	if !b.api.IsMessageToMe(msg) {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if !b.isPublicMethod(msg.Command()) {
		if !b.isAdmin(msg) {
			return false, b.ReplyOne(msg, NeedFormat{Message: onlyAdminsMessage})
		}
		return false, b.processAdminActions(ctx, msg)
	}
	return b.processPublicActions(ctx, msg)
}

func (b *Bot) processPublicActions(ctx context.Context, msg tgbotapi.Message) (del bool, err error) {
	b.log.Debug("public action", zap.String("command", msg.Command()))
	del = true
	switch msg.Command() {
	case "report":
		del, err = b.report(ctx, msg)
	case "subscribe":
		err = b.subscribe(ctx, msg)
	case "unsubscribe":
		err = b.unsubscribe(ctx, msg)
	default:
		err = fmt.Errorf("%w: %s", ErrNoSuchCommand, msg.Command())
		del = false
	}
	return
}

func (b *Bot) processAdminActions(ctx context.Context, msg tgbotapi.Message) error {
	b.log.Debug("admin action", zap.String("command", msg.Command()))
	switch msg.Command() {
	case "ban":
		return b.kick(ctx, msg, forever)
	case "kick":
		return b.kick(ctx, msg, day)
	case "warn":
		return b.warn(ctx, msg, true)
	case "unwarn":
		return b.warn(ctx, msg, false)
	default:
		return fmt.Errorf("%w: %s", ErrNoSuchCommand, msg.Command())
	}
}
