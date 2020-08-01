package bot

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/zap"
)

var funcs = template.FuncMap{
	"formatUser": func(user *tgbotapi.User) string {
		return fmt.Sprintf("[%s](tg://user?id=%d)", user.String(), user.ID)
	},
	"div100": func(num int64) string {
		return strings.TrimPrefix(strconv.FormatInt(num, 10), "-100")
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
		f.after = after
	}
}

func AddBefore(before BeforeFunc) formatterOption {
	return func(f *Formatter) {
		f.before = before
	}
}

type Formatter struct {
	log      *zap.Logger
	api      TelegramAPI
	baseChat tgbotapi.BaseChat

	after  AfterFunc
	before BeforeFunc
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
	msg := &tgbotapi.MessageConfig{
		BaseChat:  f.baseChat,
		Text:      need.Message,
		ParseMode: "markdown",
	}
	if need.FormatParams != nil {
		text, err := format(need)
		if err != nil {
			return err
		}
		msg.Text = text
	}
	if f.before != nil {
		f.before(msg)
	}
	message, err := f.api.Send(msg)
	if err != nil {
		f.log.Error("Send message",
			zap.Error(err),
			zap.Int64("chat_id", f.baseChat.ChatID),
		)
		return err
	}
	if f.after != nil {
		f.after(message)
	}
	return nil
}
