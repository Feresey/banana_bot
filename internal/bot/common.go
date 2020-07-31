package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Feresey/banana_bot/internal/db"
)

func personFromMessage(msg tgbotapi.Message) *db.Person {
	return &db.Person{
		ChatID: msg.Chat.ID,
		UserID: int64(msg.From.ID),
	}
}

func (b *Bot) ReplyOne(forward tgbotapi.Message, need NeedFormat) error {
	return NewFormatter(
		b.api, tgbotapi.BaseChat{ChatID: forward.Chat.ID, ReplyToMessageID: forward.MessageID},
	).Format(need)
}

func (b *Bot) ToChat(chatID int64, need NeedFormat) error {
	return NewFormatter(
		b.api, tgbotapi.BaseChat{ChatID: chatID},
	).Format(need)
}
