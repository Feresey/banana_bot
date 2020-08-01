package bot

import (
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

	call := commandMesage("/start@command", "private")

	for _, msg := range startMessages {
		am.EXPECT().Send(&tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: call.Chat.ID,
			},
			Text: msg.Message,
		}).Return(tgbotapi.Message{}, nil).Times(1)
	}

	err := bot.processMessage(call)
	require.NoError(t, err)
}

func TestStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	am := NewMockTelegramAPI(ctrl)
	bot.api = am

	call := commandMesage("/stop@command", "private")
	am.EXPECT().Send(&tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           call.Chat.ID,
			ReplyToMessageID: call.MessageID,
		},
		Text: mustFormat(t, NeedFormat{Message: stopMessage, FormatParams: call.From}),
	})

	err := bot.processMessage(call)
	require.NoError(t, err)
}

func TestDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	am := NewMockTelegramAPI(ctrl)
	bot.api = am

	call := commandMesage("/lol_none@command", "private")

	am.EXPECT().Send(&tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           call.Chat.ID,
			ReplyToMessageID: call.MessageID,
		},
		Text: todoMessage,
	})

	err := bot.processMessage(call)
	require.NoError(t, err)
}
