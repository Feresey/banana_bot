package bot

import (
	"context"
	"time"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/zap"
)

var (
	subscribeMessage   = "Лайк, подписка"
	unsubscribeMessage = "Дизлайк, отписка"
)

func formatReport(msg *tgbotapi.Message) NeedFormat {
	return NeedFormat{
		Message:      "Вас призывают в чат\n{{formatChatMessage .Chat .MessageID}}",
		FormatParams: msg,
	}
}

func (b *Bot) deleteAfter(wait time.Duration) AfterFunc {
	return func(msg *tgbotapi.Message) {
		time.Sleep(wait)
		_, err := b.api.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    msg.Chat.ID,
			MessageID: msg.MessageID,
		})
		if err != nil {
			b.log.Error("Delete message by timeout", zap.Error(err))
			return
		}
		b.log.Debug("Delete message by timeout")
	}
}

func formatCalled() NeedFormat { return NeedFormat{Message: "Админы призваны!"} }

func (b *Bot) report(ctx context.Context, msg *tgbotapi.Message) (bool, error) {
	subscribed, err := b.db.Report(ctx, msg.Chat.ID)
	if err != nil {
		return false, err
	}

	message := make(chan *tgbotapi.Message)

	err = b.ToChat(
		msg.Chat.ID,
		formatCalled(),
		AddAfter(b.deleteAfter(time.Minute)),
		AddAfter(func(msg *tgbotapi.Message) { message <- msg }),
	)
	if err != nil {
		b.log.Warn("Report message", zap.Error(err))
	}

	for _, subscriber := range subscribed {
		if err := b.ToChat(subscriber, formatReport(<-message)); err != nil {
			b.log.Error("Send report to subscriber", zap.Error(err))
		}
	}
	// ваще хз, удалять такое или нет
	return false, nil
}

func (b *Bot) subscribe(ctx context.Context, msg *tgbotapi.Message) error {
	if err := b.db.Subscribe(ctx, personFromMessage(msg)); err != nil {
		return err
	}
	return b.ReplyOne(msg, NeedFormat{Message: subscribeMessage}, AddAfter(b.deleteAfter(5*time.Second)))
}

func (b *Bot) unsubscribe(ctx context.Context, msg *tgbotapi.Message) error {
	if err := b.db.Unsubscribe(ctx, personFromMessage(msg)); err != nil {
		return err
	}
	return b.ReplyOne(msg, NeedFormat{Message: unsubscribeMessage}, AddAfter(b.deleteAfter(5*time.Second)))
}
