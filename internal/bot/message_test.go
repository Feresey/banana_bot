package bot

import (
	"testing"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// func send(msg tgbotapi.Message) {
// 	done := make(chan struct{})
// 	go func() {
// 		updates <- tgbotapi.Update{Message: &msg}
// 	}()
// 	<-done
// }

func commandMesage(cmd, group string) *tgbotapi.Message {
	return &tgbotapi.Message{
		Entities: &[]tgbotapi.MessageEntity{{
			Length: len(cmd),
			Offset: 0,
			Type:   "bot_command",
		}},
		Text: cmd,
		Chat: &tgbotapi.Chat{
			Type: group,
			ID:   12312312312123123,
		},
		MessageID: 4422,
		From: &tgbotapi.User{
			ID:           1111,
			FirstName:    "John",
			LastName:     "Smith",
			UserName:     "user",
			IsBot:        false,
			LanguageCode: "en-US",
		},
	}
}

func mustFormat(t *testing.T, need NeedFormat) string {
	out, err := format(need)
	require.NoError(t, err)
	return out
}

func TestProtect(t *testing.T) {
	t.Run("not admin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)

		msg := commandMesage("/ban@bot", "group")
		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "really not admin",
			},
			nil,
		).Times(1)

		api.EXPECT().Send(
			&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID:           msg.Chat.ID,
					ReplyToMessageID: msg.MessageID,
				},
				Text: onlyAdminsMessage,
			}).Return(nil, nil).Times(1)

		bot.api = api
		err := bot.processMessage(msg)
		require.NoError(t, err)
	})

	t.Run("sudo mazo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)

		msg := commandMesage("/ban@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{From: msg.From}

		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		api.EXPECT().Send(
			&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID:           msg.Chat.ID,
					ReplyToMessageID: msg.MessageID,
				},
				Text: mustFormat(t,
					NeedFormat{
						Message:      protectSelfMessage,
						FormatParams: msg.From,
					}),
			},
		).Return(nil, nil).Times(1)

		bot.api = api
		err := bot.processMessage(msg)
		require.NoError(t, err)
	})

	t.Run("me", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)

		msg := commandMesage("/ban@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{From: &tgbotapi.User{
			ID:    bot.me.ID,
			IsBot: true,
		}}

		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		api.EXPECT().Send(
			&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID:           msg.Chat.ID,
					ReplyToMessageID: msg.MessageID,
				},
				Text: mustFormat(t, NeedFormat{Message: protectMeMessage}),
			},
		).Return(nil, nil).Times(1)

		bot.api = api
		err := bot.processMessage(msg)
		require.NoError(t, err)
	})

	t.Run("god", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)

		msg := commandMesage("/ban@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{From: &tgbotapi.User{
			ID:    godID,
			IsBot: false,
		}}

		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		api.EXPECT().Send(
			&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID:           msg.Chat.ID,
					ReplyToMessageID: msg.MessageID,
				},
				Text: mustFormat(t, NeedFormat{Message: protectGodMessage}),
			},
		).Return(nil, nil).Times(1)

		bot.api = api
		err := bot.processMessage(msg)
		require.NoError(t, err)
	})
}

func TestWarn(t *testing.T) {
	t.Run("first warn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)
		dbm := NewMockDatabase(ctrl)
		bot.api = api
		bot.db = dbm

		msg := commandMesage("/warn@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{
			MessageID: 31333,
			From: &tgbotapi.User{
				ID:       99910,
				UserName: "caller",
			},
			Chat: msg.Chat,
		}
		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		target := personFromMessage(msg.ReplyToMessage)

		dbm.EXPECT().
			Warn(gomock.Any(), target, true).Return(int64(1), nil)

		api.EXPECT().
			RestrictChatMember(bot.limitation(target, 1)).
			Return(nil, nil)
		api.EXPECT().
			Send(&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID: msg.Chat.ID,
				},
				Text: mustFormat(t, formatWarn(msg.ReplyToMessage.From, 1, bot.c.MaxWarn)),
			})
		api.EXPECT().
			DeleteMessage(tgbotapi.DeleteMessageConfig{
				ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.MessageID}).Return(nil, nil)

		err := bot.processMessage(msg)
		require.NoError(t, err)
	})

	t.Run("unwarn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)
		dbm := NewMockDatabase(ctrl)
		bot.api = api
		bot.db = dbm

		msg := commandMesage("/unwarn@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{
			MessageID: 31333,
			From: &tgbotapi.User{
				ID:       99910,
				UserName: "caller",
			},
			Chat: msg.Chat,
		}
		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		target := personFromMessage(msg.ReplyToMessage)

		dbm.EXPECT().
			Warn(gomock.Any(), target, false).Return(int64(0), nil).Times(1)

		api.EXPECT().
			Send(&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID: msg.Chat.ID,
				},
				Text: mustFormat(t, formatUnwarn(msg.ReplyToMessage.From, 0)),
			})

		err := bot.processMessage(msg)
		require.NoError(t, err)
	})

	t.Run("last warn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)
		dbm := NewMockDatabase(ctrl)
		bot.api = api
		bot.db = dbm

		msg := commandMesage("/warn@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{
			MessageID: 31333,
			From: &tgbotapi.User{
				ID:       99910,
				UserName: "caller",
			},
			Chat: msg.Chat,
		}
		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		target := personFromMessage(msg.ReplyToMessage)

		dbm.EXPECT().
			Warn(gomock.Any(), target, true).Return(bot.c.MaxWarn, nil)

		api.EXPECT().
			RestrictChatMember(bot.limitation(target, bot.c.MaxWarn)).
			Return(nil, nil)
		api.EXPECT().
			Send(&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID: msg.Chat.ID,
				},
				Text: mustFormat(t, formatLastWarn(msg.ReplyToMessage.From)),
			})
		api.EXPECT().
			DeleteMessage(tgbotapi.DeleteMessageConfig{
				ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.MessageID}).Return(nil, nil)

		err := bot.processMessage(msg)
		require.NoError(t, err)
	})

	t.Run("kick", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)
		dbm := NewMockDatabase(ctrl)
		bot.api = api
		bot.db = dbm

		msg := commandMesage("/warn@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{
			MessageID: 31333,
			From: &tgbotapi.User{
				ID:       99910,
				UserName: "caller",
			},
			Chat: msg.Chat,
		}
		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		target := personFromMessage(msg.ReplyToMessage)

		dbm.EXPECT().
			Warn(gomock.Any(), target, true).Return(bot.c.MaxWarn+1, nil)
		api.EXPECT().
			RestrictChatMember(bot.limitation(target, bot.c.MaxWarn+1)).
			Return(nil, nil)
		api.EXPECT().
			KickChatMember(getKickConfig(target, forever))
		api.EXPECT().
			DeleteMessage(tgbotapi.DeleteMessageConfig{
				ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.MessageID}).Return(nil, nil)
		api.EXPECT().
			Send(&tgbotapi.MessageConfig{BaseChat: tgbotapi.BaseChat{
				ChatID: msg.Chat.ID,
			},
				Text:      mustFormat(t, formatKick(msg.ReplyToMessage.From, forever)),
				ParseMode: "markdown",
			})

		err := bot.processMessage(msg)
		require.NoError(t, err)
	})
}

func TestKick(t *testing.T) {
	t.Run("kick", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)
		dbm := NewMockDatabase(ctrl)
		bot.api = api
		bot.db = dbm

		msg := commandMesage("/kick@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{
			MessageID: 31333,
			From: &tgbotapi.User{
				ID:       99910,
				UserName: "caller",
			},
			Chat: msg.Chat,
		}
		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		target := personFromMessage(msg.ReplyToMessage)

		api.EXPECT().
			KickChatMember(getKickConfig(target, day)).
			Return(nil, nil)
		api.EXPECT().
			DeleteMessage(tgbotapi.DeleteMessageConfig{
				ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.MessageID}).Return(nil, nil)
		api.EXPECT().
			Send(&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID: msg.Chat.ID,
				},
				Text: mustFormat(t, formatKick(msg.ReplyToMessage.From, day)),
			})

		err := bot.processMessage(msg)
		require.NoError(t, err)
	})

	t.Run("ban", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		api := NewMockTelegramAPI(ctrl)
		dbm := NewMockDatabase(ctrl)
		bot.api = api
		bot.db = dbm

		msg := commandMesage("/ban@bot", "group")
		msg.ReplyToMessage = &tgbotapi.Message{
			MessageID: 31333,
			From: &tgbotapi.User{
				ID:       99910,
				UserName: "caller",
			},
			Chat: msg.Chat,
		}
		api.EXPECT().IsMessageToMe(msg).Return(true).Times(1)
		api.EXPECT().GetChatMember(
			tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: msg.From.ID,
			}).Return(
			&tgbotapi.ChatMember{
				User:   msg.From,
				Status: "administrator",
			},
			nil,
		).Times(1)

		target := personFromMessage(msg.ReplyToMessage)

		api.EXPECT().
			KickChatMember(getKickConfig(target, forever)).
			Return(nil, nil)
		api.EXPECT().
			DeleteMessage(tgbotapi.DeleteMessageConfig{
				ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.MessageID}).Return(nil, nil)
		api.EXPECT().
			Send(&tgbotapi.MessageConfig{
				ParseMode: "markdown",
				BaseChat: tgbotapi.BaseChat{
					ChatID: msg.Chat.ID,
				},
				Text: mustFormat(t, formatKick(msg.ReplyToMessage.From, forever)),
			})

		err := bot.processMessage(msg)
		require.NoError(t, err)
	})
}
