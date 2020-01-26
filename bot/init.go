package bot

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Feresey/banana_bot/db"
	"github.com/Feresey/banana_bot/logging"
	"github.com/Feresey/banana_bot/model"
)

var (
	bot     *tgbotapi.BotAPI
	log     = logging.NewLogger("Bot")
	debug   = false
	maxWarn = 5
)

func updateAdmins() {
	updateAllAdmins()
	for range time.Tick(time.Hour) {
		updateAllAdmins()
	}
}

// Start : initialize a bot
func Start(token string, d bool) error {
	debug = d

	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		if token == "" {
			flag.Usage()
			log.Fatal("TOKEN???")
		} else {
			log.Fatal(err)
		}
	}
	if err != nil {
		panic(err)
	}
	bot = b
	log.Infof("Successfully connected on %s", bot.Self.UserName)

	err = db.Connect(log)
	if err != nil {
		return err
	}
	log.Info("DB connected")

	go updateAdmins()
	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint
		db.Shutdown()
		log.Info("Bot closed." + time.Now().Format(time.Stamp))
		os.Exit(0)
	}()

	return nil
}

// KeepOn : start messaging and block main function
func KeepOn() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Unable get updates", err)
	}

	for update := range updates {
		msg := update.Message
		if msg != nil {
			go processMessage(&model.Message{Message: msg})
		}
	}
}
