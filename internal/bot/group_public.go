package bot

import (
	"context"

	"github.com/Feresey/banana_bot/internal/format"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) report(ctx context.Context, msg tgbotapi.Message) (bool, error) {
	subscribed, err := b.db.Report(ctx, msg.Chat.ID)
	if err != nil {
		return false, err
	}

	for _, subscriber := range subscribed {
		if err := b.ToChat(
			subscriber,
			format.NeedFormat{
				Message:      "Вас призывают в чат {{.Title}}.",
				FormatParams: msg.Chat,
			}); err != nil {
			return false, err
		}
	}
	return true, b.ToChat(msg.Chat.ID, format.NeedFormat{Message: "Админы призваны!"})
}

func (b *Bot) subscribe(ctx context.Context, msg tgbotapi.Message) error {
	if err := b.db.Subscribe(ctx, personFromMessage(msg)); err != nil {
		return err
	}
	return b.ReplyOne(msg, format.NeedFormat{Message: "Лайк, подписка"})
}

func (b *Bot) unsubscribe(ctx context.Context, msg tgbotapi.Message) error {
	if err := b.db.Unsubscribe(ctx, personFromMessage(msg)); err != nil {
		return err
	}
	return b.ReplyOne(msg, format.NeedFormat{Message: "Дизайк, отписка"})
}
