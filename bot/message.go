package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Feresey/banana_bot/format"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

var (
	ErrWTF           = errors.New("what? not group, not private. what is it???")
	ErrNoSuchCommand = errors.New("no such command")
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
	}
	return ErrWTF
}

func (b *Bot) groupMessage(msg tgbotapi.Message) (del bool, err error) {
	if !msg.IsCommand() {
		// Обычное сообщение, лог ненужон
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if !b.isPublicMethod(msg.Command()) {
		if !b.isAdmin(msg) {
			return false, b.ReplyOne(msg, format.NeedFormat{Message: "Только админам можно"})
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
