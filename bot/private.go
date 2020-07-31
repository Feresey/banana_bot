package bot

import (
	"time"

	"github.com/Feresey/banana_bot/format"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) privateMessage(msg tgbotapi.Message) error {
	cmd := msg.Command()

	switch cmd {
	case "start":
		messages := []format.NeedFormat{
			{Message: "Приветствую, кожаный мешок!"},
			{Message: "Я буду отсылать тебе сообщения о репортах."},
			{Message: "Если тебе надоест этот \"спам\", то просто удали чат со мной."},
			{Message: "(всё гениальное просто, да)"},
		}
		formatter := format.New(
			b.api, tgbotapi.BaseChat{ChatID: msg.Chat.ID},
			format.AddAfter(func(tgbotapi.Message) { time.Sleep(time.Second) }),
		)

		for _, msg := range messages {
			if err := formatter.Format(msg); err != nil {
				return err
			}
		}
	case "stop":
		return b.ReplyOne(msg,
			format.NeedFormat{
				Message: "Хорошая попытка, {{formatUser .}}, " +
					"но от меня так просто не избавиться!",
				FormatParams: msg.From,
			},
		)
	default:
		return b.ReplyOne(msg, format.NeedFormat{
			Message: "Соре, я не умею ничего в личном чате. Хз, может тут будет статистика.",
		})
	}
	return nil
}
