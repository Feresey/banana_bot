package service

import (
	"sync"
	"time"

	"github.com/Feresey/banana_bot/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hokaccha/go-prettyjson"
)

var (
	log   *logging.Logger
	bot   *tgbotapi.BotAPI
	Debug = false

	// map[chatID][]userID
	AdminList = map[int64][]tgbotapi.ChatMember{}
	mu        = &sync.RWMutex{}
)

func Init(logger *logging.Logger, bott *tgbotapi.BotAPI) {
	log = logger.Child("Service")
	bot = bott
	getAdmins()
	go func() {
		for range time.Tick(time.Hour) {
			getAdmins()
		}
	}()
}

func StartMessaging(u tgbotapi.UpdateConfig) {
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Unable get updates", err)
	}

	for update := range updates {
		go ProcessMessage(update)
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

	// if msg.LeftChatMember != nil {
	// 	reply := tgbotapi.NewMessage(msg.Chat.ID, "F")
	// 	sendMsg(reply)
	// }
}
