package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hokaccha/go-prettyjson"
)

func sendMsg(msg tgbotapi.MessageConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Error("Unable to send message:", err)
	}
}

func ProcessMessage(update tgbotapi.Update) {
	msg := update.Message
	if msg == nil { // ignore any non-Message Updates
		return
	}
	if Debug {
		data, _ := prettyjson.Marshal(update.Message)
		log.Info(string(data))
	} else {
		log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}

	cmd := msg.Command()
	switch cmd {
	case "report":
		makeReportAdmins(msg)
	case "ban":
		makeBan(msg)
	case "warn":
		makeWarn(msg, true)
	case "unwarn":
		makeWarn(msg, false)
	case "start":
		startPrivateChat(msg)
	}

	if msg.LeftChatMember != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "F")
		sendMsg(reply)
	}
}
