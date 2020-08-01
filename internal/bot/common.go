package bot

import (
	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/Feresey/banana_bot/internal/db"
)

func personFromMessage(msg *tgbotapi.Message) *db.Person {
	return &db.Person{
		ChatID: msg.Chat.ID,
		UserID: int64(msg.From.ID),
	}
}

func (b *Bot) ReplyOne(forward *tgbotapi.Message, need NeedFormat, opts ...formatterOption) error {
	b.log.Info("Send reply", zap.Int("message_id", forward.MessageID))
	return NewFormatter(b.log,
		b.api, tgbotapi.BaseChat{ChatID: forward.Chat.ID, ReplyToMessageID: forward.MessageID},
		opts...,
	).Format(need)
}

func (b *Bot) ToChat(chatID int64, need NeedFormat, opts ...formatterOption) error {
	b.log.Info("Send to chat", zap.Int64("chat_id", chatID))
	return NewFormatter(b.log,
		b.api, tgbotapi.BaseChat{ChatID: chatID},
		opts...,
	).Format(need)
}
