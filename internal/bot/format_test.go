package bot

import (
	"errors"
	"fmt"
	"testing"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	api := NewMockTelegramAPI(ctrl)

	var (
		user = &tgbotapi.User{
			ID:        42,
			FirstName: "John",
			LastName:  "Smith",
		}
		params = map[string]interface{}{
			"Message": "really stuff",
			"User":    user,
		}
		need = NeedFormat{
			Message:      "format stuff: {{.Message}}. And user: {{formatUser .User}}",
			FormatParams: params,
		}

		want = fmt.Sprintf(
			"format stuff: %s. And user: [%s](tg://user?id=%d)",
			params["Message"],
			user.String(), user.ID)

		baseChat = tgbotapi.BaseChat{
			ChatID:           4444,
			ReplyToMessageID: 123123,
		}
	)

	api.EXPECT().Send(&tgbotapi.MessageConfig{
		BaseChat: baseChat,
		Text:     want,
	}).Return(nil, nil).Times(1)

	err := NewFormatter(api, baseChat).Format(need)
	require.NoError(t, err)

	wantErr := errors.New("error")
	api.EXPECT().Send(&tgbotapi.MessageConfig{
		BaseChat: baseChat,
		Text:     want,
	}).Return(nil, wantErr).Times(1)
	err = NewFormatter(api, baseChat).Format(need)
	require.EqualError(t, err, wantErr.Error())
}
