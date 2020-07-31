package bot

import (
	"html/template"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	am := NewMockTelegramAPI(ctrl)
	bot.api = am

	text := "/start@command"
	call := tgbotapi.Message{
		MessageID: 1,
		Entities: &[]tgbotapi.MessageEntity{{
			Offset: 0,
			Length: len(text),
			Type:   "bot_command",
		}},
		Text: text,
		Chat: &tgbotapi.Chat{
			ID:    123123,
			Title: "chat",
		},
	}

	for _, msg := range startMessages {
		am.EXPECT().Send(&tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: call.Chat.ID,
			},
			Text: msg.Message,
		}).Return(tgbotapi.Message{}, nil).Times(1)
	}

	err := bot.privateMessage(call)
	require.NoError(t, err)
}

func TestStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	am := NewMockTelegramAPI(ctrl)
	bot.api = am

	text := "/stop@command"
	call := tgbotapi.Message{
		MessageID: 1,
		Entities: &[]tgbotapi.MessageEntity{{
			Offset: 0,
			Length: len(text),
			Type:   "bot_command",
		}},
		Text: text,
		Chat: &tgbotapi.Chat{
			ID:    123123,
			Title: "chat",
		},
		From: &tgbotapi.User{
			ID:       42,
			UserName: "user",
		},
	}

	out := new(strings.Builder)
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(stopMessage))
	err := tmpl.Execute(out, call.From)
	require.NoError(t, err)
	want := out.String()

	am.EXPECT().Send(&tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           call.Chat.ID,
			ReplyToMessageID: call.MessageID,
		},
		Text: want,
	})

	err = bot.privateMessage(call)
	require.NoError(t, err)
}

func TestDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	am := NewMockTelegramAPI(ctrl)
	bot.api = am

	text := "/lol_none@command"
	call := tgbotapi.Message{
		MessageID: 1,
		Entities: &[]tgbotapi.MessageEntity{{
			Offset: 0,
			Length: len(text),
			Type:   "bot_command",
		}},
		Text: text,
		Chat: &tgbotapi.Chat{
			ID:    123123,
			Title: "chat",
		},
	}

	am.EXPECT().Send(&tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           call.Chat.ID,
			ReplyToMessageID: call.MessageID,
		},
		Text: todoMessage,
	})

	err := bot.privateMessage(call)
	require.NoError(t, err)
}
