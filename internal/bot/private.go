package bot

import (
	"time"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
)

var (
	startMessages = []NeedFormat{
		{Message: "Приветствую, кожаный мешок!"},
		{Message: "Я буду отсылать тебе сообщения о репортах."},
		{Message: "Если тебе надоест этот \"спам\", то просто удали чат со мной."},
		{Message: "(всё гениальное просто, да)"},
	}

	stopMessage = "Хорошая попытка, {{formatUser .}}, " +
		"но от меня так просто не избавиться!"

	todoMessage = "Соре, я не умею ничего в личном чате. Хз, может тут будет статистика."
)

func (b *Bot) privateMessage(msg *tgbotapi.Message) error {
	cmd := msg.Command()

	switch cmd {
	case "start":
		formatter := NewFormatter(b.log,
			b.api, tgbotapi.BaseChat{ChatID: msg.Chat.ID},
			AddAfter(func(*tgbotapi.Message) { time.Sleep(b.c.ResponseSleep) }),
		)

		for _, msg := range startMessages {
			if err := formatter.Format(msg); err != nil {
				return err
			}
		}
	case "stop":
		return b.ReplyOne(msg,
			NeedFormat{
				Message:      stopMessage,
				FormatParams: msg.From,
			},
		)
	default:
		return b.ReplyOne(msg, NeedFormat{
			Message: todoMessage,
		})
	}
	return nil
}
