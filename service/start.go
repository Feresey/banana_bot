package service

import (
	"github.com/Feresey/banana_bot/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	log   *logging.Logger
	bot   *tgbotapi.BotAPI
	Debug = false
)

func Init(logger *logging.Logger, bott *tgbotapi.BotAPI) {
	log = logger.Child("Service")
	bot = bott
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
