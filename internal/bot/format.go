package bot

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	formatChat        = `[%s](https://t.me/c/%s)`
	formatChatMessage = `[%s](https://t.me/c/%s/%d)`
)

func trim100(num int64) string {
	return strings.TrimPrefix(strconv.FormatInt(num, 10), "-100")
}

var quote = map[rune]struct{}{
	'[': {},
	']': {},
	')': {},
	'(': {},
}

func quoteMD(s string) (res string) {
	for _, c := range s {
		if _, ok := quote[c]; ok {
			c = ' '
		}
		res += string(c)
	}
	return
}

var funcs = template.FuncMap{
	"formatUser": func(user *tgbotapi.User) string {
		return fmt.Sprintf("[%s](tg://user?id=%d)", user.String(), user.ID)
	},
	"formatChat": func(chat *tgbotapi.Chat) string {
		title := chat.Title
		if title == "" {
			title = chat.UserName
		}
		return fmt.Sprintf(formatChat, quoteMD(title), trim100(chat.ID))
	},
	"formatChatMessage": func(chat *tgbotapi.Chat, messageID int) string {
		return fmt.Sprintf(formatChatMessage, quoteMD(chat.Title), trim100(chat.ID), messageID)
	},
}

type NeedFormat struct {
	// i18n?
	Message      string
	FormatParams interface{}
}

type FormatOptions struct {
	tgbotapi.BaseChat
}

type formatterOption func(f *Formatter)

type AfterFunc func(*tgbotapi.Message)
type BeforeFunc func(*tgbotapi.MessageConfig)

func AddAfter(after AfterFunc) formatterOption {
	return func(f *Formatter) {
		f.afterFuncs = append(f.afterFuncs, after)
	}
}

func AddBefore(before BeforeFunc) formatterOption {
	return func(f *Formatter) {
		f.beforeFuncs = append(f.beforeFuncs, before)
	}
}

type Formatter struct {
	log      *zap.Logger
	api      TelegramAPI
	baseChat tgbotapi.BaseChat

	afterFuncs  []AfterFunc
	beforeFuncs []BeforeFunc
}

func NewFormatter(
	log *zap.Logger,
	api TelegramAPI,
	baseChat tgbotapi.BaseChat,
	opts ...formatterOption,
) *Formatter {
	f := &Formatter{
		log:      log.Named("format"),
		api:      api,
		baseChat: baseChat,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func format(format NeedFormat) (string, error) {
	tmpl, err := template.New("").Funcs(funcs).Parse(format.Message)
	if err != nil {
		return "", err
	}

	out := new(strings.Builder)
	if err = tmpl.Execute(out, format.FormatParams); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (f *Formatter) Format(need NeedFormat) error {
	send := &tgbotapi.MessageConfig{
		BaseChat:  f.baseChat,
		Text:      need.Message,
		ParseMode: "markdown",
	}
	if need.FormatParams != nil {
		text, err := format(need)
		if err != nil {
			return err
		}
		send.Text = text
	}
	for _, fn := range f.beforeFuncs {
		fn(send)
	}
	message, err := f.api.Send(send)
	if err != nil {
		f.log.Error("Send message",
			zap.Error(err),
			zap.Int64("chat_id", f.baseChat.ChatID),
		)
		return err
	}
	for _, fn := range f.afterFuncs {
		go fn(message)
	}
	return nil
}
