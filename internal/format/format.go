package format

import (
	"fmt"
	"html/template"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var funcs = template.FuncMap{
	"formatUser": func(user tgbotapi.User) string {
		return fmt.Sprintf("[%s](tg://user?id=%d)", user.UserName, user.ID)
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

type AfterFunc func(tgbotapi.Message)
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
	api      *tgbotapi.BotAPI
	baseChat tgbotapi.BaseChat

	after  AfterFunc
	before BeforeFunc
}

func New(api *tgbotapi.BotAPI, baseChat tgbotapi.BaseChat, opts ...formatterOption) *Formatter {
	f := &Formatter{api: api}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *Formatter) Format(format NeedFormat) error {
	msg := &tgbotapi.MessageConfig{
		BaseChat: f.baseChat,
	}
	tmpl, err := template.New("").Funcs(funcs).Parse(format.Message)
	if err != nil {
		return err
	}

	out := new(strings.Builder)
	if err = tmpl.Execute(out, format.FormatParams); err != nil {
		return err
	}

	msg.Text = out.String()
	if f.before != nil {
		f.before(msg)
	}
	message, err := f.api.Send(msg)
	if err != nil {
		return err
	}
	if f.after != nil {
		f.after(message)
	}
	return nil
}
