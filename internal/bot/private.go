package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/zap"
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

func (b *Bot) pushLogs() error {
	b.log.Debug("Send log file", zap.String("filename", b.c.LogFile))
	file := tgbotapi.NewDocumentUpload(godID, b.c.LogFile)
	_, err := b.api.Send(file)
	return err
}

// TODO inline action
func (b *Bot) handleText(msg *tgbotapi.Message) error {
	switch msg.Text {
	case "logs":
		if msg.From.ID != godID {
			return fmt.Errorf("you are not god")
		}

		return b.pushLogs()
	default:
		return nil
	}
}

func (b *Bot) privateMessage(msg *tgbotapi.Message) error {
	if !msg.IsCommand() {
		return b.handleText(msg)
	}

	switch msg.Command() {
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
	case "subscribe":
		return b.subscribePrivate(msg)
	default:
		return b.ReplyOne(msg, NeedFormat{
			Message: todoMessage,
		})
	}
	return nil
}

func (b *Bot) subscribePrivate(msg *tgbotapi.Message) error {
	found, err := b.findChats(msg.From.ID)
	if err != nil {
		return err
	}

	reply := formatChatsNotFound()
	if len(found) != 0 {
		reply = formatChatsFound(found)
	}
	if err := b.ToChat(msg.Chat.ID, reply); err != nil {
		return fmt.Errorf("reply with found chats: %w", err)
	}
	if len(found) == 0 {
		return nil
	}
	// TODO : incline query
	return nil
}

func formatChatsNotFound() NeedFormat {
	return NeedFormat{
		Message: "Я не нашёл тебя в чатах со мной. Бывает.",
	}
}

func formatChatsFound(chats []*tgbotapi.Chat) NeedFormat {
	return NeedFormat{
		Message: `Я нашёл тебя в этих чатах:
{{range $idx, $value := .}}
{{$idx}}) {{$value.Title}}.
{{end}}

На какие хочешь подписаться?
Ответь списком чисел`,
		FormatParams: chats,
	}
}

func (b *Bot) findChats(userID int) ([]*tgbotapi.Chat, error) {
	getChatsCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	chats, err := b.db.GetMyChats(getChatsCtx)
	if err != nil {
		return nil, err
	}

	found := make([]*tgbotapi.Chat, 0, len(chats))

	for _, chatID := range chats {
		member, err := b.api.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		})
		if err != nil {
			b.log.Error("Get chat member", zap.Error(err))
			continue
		}

		if !member.IsMember() {
			continue
		}

		chat, err := b.api.GetChat(tgbotapi.ChatConfig{ChatID: chatID})
		if err != nil {
			b.log.Warn("Get chat by id", zap.Int64("chat_id", chatID), zap.Error(err))
			continue
		}

		found = append(found, chat)
	}

	return found, nil
}
