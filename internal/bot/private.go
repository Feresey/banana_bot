package bot

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Feresey/banana_bot/internal/db"
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
			AddBefore(func(*tgbotapi.MessageConfig) { time.Sleep(b.c.ResponseSleep) }),
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
		return b.subscribeInPrivate(msg)
	default:
		return b.ReplyOne(msg, NeedFormat{
			Message: todoMessage,
		})
	}
	return nil
}

type subscribeUser struct {
	b *Bot

	UserID    int
	MessageID int
	mu        sync.Mutex
	Chats     map[int64]*tgbotapi.Chat
}

func (b *Bot) newSubscriber(msg *tgbotapi.Message, chats map[int64]*tgbotapi.Chat) *subscribeUser {
	return &subscribeUser{
		b:         b,
		Chats:     chats,
		MessageID: msg.MessageID,
		UserID:    msg.From.ID,
	}
}

func (s *subscribeUser) subscribe(upd *tgbotapi.CallbackQuery) {
	chatID, err := strconv.ParseInt(upd.Data, 10, 64)
	if err != nil {
		s.b.log.Error("Could not parse chat ID from callback",
			zap.Error(err), zap.String("data", upd.Data))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.Chats[chatID]
	if !exists {
		s.b.log.Error("Попытка наебать detected",
			zap.Int("пидор_id", s.UserID),
			zap.Reflect("chats", s.Chats),
			zap.Int64("chat_id", chatID),
		)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err = s.b.db.Subscribe(ctx, &db.Person{ChatID: chatID, UserID: int64(s.UserID)})
	if err != nil {
		s.b.log.Error("Try to subscribe by callback", zap.Error(err))
		_, _ = s.b.api.AnswerCallbackQuery(tgbotapi.CallbackConfig{
			CallbackQueryID: upd.ID,
			ShowAlert:       true,
			Text:            err.Error(),
		})
	}
	_, _ = s.b.api.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: upd.ID,
		ShowAlert:       true,
		Text:            "подписка успешна",
	})
}

func (b *Bot) subscribeInPrivate(msg *tgbotapi.Message) error {
	found, err := b.findChats(msg.From.ID)
	if err != nil {
		return err
	}

	reply := formatChatsNotFound()
	if len(found) != 0 {
		reply = formatChatsFound()
	}

	msgIdCallback := make(chan int)
	err = b.ToChat(msg.Chat.ID,
		reply,
		AddAfter(func(msg *tgbotapi.Message) { msgIdCallback <- msg.MessageID }),
	)
	if err != nil {
		return fmt.Errorf("reply with found chats: %w", err)
	}
	if len(found) == 0 {
		return nil
	}

	foundChats := make([]tgbotapi.InlineKeyboardButton, 0, len(found))
	for _, foundChat := range found {
		title := foundChat.Title
		if title == "" {
			title = foundChat.UserName
		}
		button := tgbotapi.NewInlineKeyboardButtonData(title, strconv.FormatInt(foundChat.ID, 10))
		foundChats = append(foundChats, button)
	}

	id := <-msgIdCallback
	keys := tgbotapi.NewInlineKeyboardMarkup(foundChats)
	edit := tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, id, keys)
	if _, err := b.api.Send(edit); err != nil {
		return err
	}

	sub := b.newSubscriber(msg, found)
	b.cs.AddCallback(id, sub.subscribe, time.Minute)

	return nil
}

func formatChatsNotFound() NeedFormat {
	return NeedFormat{Message: "Я не нашёл тебя в чатах со мной. Бывает."}
}

func formatChatsFound() NeedFormat {
	return NeedFormat{Message: "Я нашёл тебя в чатах.На какие хочешь подписаться?"}
}

func (b *Bot) findChats(userID int) (map[int64]*tgbotapi.Chat, error) {
	getChatsCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	chats, err := b.db.GetMyChats(getChatsCtx)
	if err != nil {
		return nil, err
	}

	found := make(map[int64]*tgbotapi.Chat, len(chats))

	for _, chatID := range chats {
		member, err := b.api.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		})
		if err != nil {
			b.log.Error("Get chat member", zap.Error(err))
			continue
		}

		// banned
		if member.HasLeft() || member.WasKicked() {
			continue
		}

		chat, err := b.api.GetChat(tgbotapi.ChatConfig{ChatID: chatID})
		if err != nil {
			b.log.Warn("Get chat by id", zap.Int64("chat_id", chatID), zap.Error(err))
			continue
		}

		found[chatID] = chat
	}

	return found, nil
}
