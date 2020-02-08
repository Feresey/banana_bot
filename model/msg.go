package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type (
	// Reply : set chat id from message
	Reply struct {
		tgbotapi.MessageConfig
	}
	// Message : my methods
	Message struct {
		*tgbotapi.Message
	}
)

// NewReply : create reply from message
func NewReply(msg Message) *Reply {
	return &Reply{
		tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: msg.Chat.ID,
			},
		},
	}
}

// Reply : set Parent message
func (M *Reply) Reply(m Message) {
	M.ReplyToMessageID = m.MessageID
}
