package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Feresey/bot-tg/db"
	"github.com/Feresey/bot-tg/logging"
	"github.com/Feresey/bot-tg/service"
)

func main() {
	flag.BoolVar(&service.Debug, "debug", false, "print all message data")
	flag.Parse()
	log := logging.NewLogger("Bot")
	token := os.Getenv("TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		if token == "" {
			flag.Usage()
			log.Fatal("TOKEN???")
		} else {
			log.Fatal(err)
		}
	}

	db.Connect(log)
	service.Init(log, bot)
	go notifyExit(log)

	log.Infof("Successfully connected on %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	service.StartMessaging(u)
}

func notifyExit(log *logging.Logger) {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt)
	signal.Notify(sigint, syscall.SIGTERM)

	<-sigint
	log.Info("\nBot closed." + time.Now().String())
	os.Exit(0)
}
