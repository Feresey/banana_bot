package bot

import (
	"context"

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
			NeedFormat{
				Message: "Вас призывают в чат {{.Chat.Title}}. https://t.me/c/{{.Chat.ID % 1000000000000}}/{{.MessageID}}",

				FormatParams: msg,
			}); err != nil {
			return false, err
		}
	}
	return true, b.ToChat(msg.Chat.ID, NeedFormat{Message: "Админы призваны!"})
}

func (b *Bot) subscribe(ctx context.Context, msg tgbotapi.Message) error {
	if err := b.db.Subscribe(ctx, personFromMessage(msg)); err != nil {
		return err
	}
	return b.ReplyOne(msg, NeedFormat{Message: "Лайк, подписка"})
}

func (b *Bot) unsubscribe(ctx context.Context, msg tgbotapi.Message) error {
	if err := b.db.Unsubscribe(ctx, personFromMessage(msg)); err != nil {
		return err
	}
	return b.ReplyOne(msg, NeedFormat{Message: "Дизайк, отписка"})
}
