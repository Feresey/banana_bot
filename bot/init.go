package bot

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Feresey/banana_bot/db"
	"github.com/Feresey/banana_bot/logging"
	"github.com/Feresey/banana_bot/model"
)

// Bot : my telegram bot
type Bot struct {
	*tgbotapi.BotAPI
	log     *logging.Logger
	debug   bool
	maxWarn int
}

// NewBot : creates a new bot with token
func NewBot(token string, debug bool) *Bot {
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

	bot := &Bot{
		BotAPI:  b,
		debug:   debug,
		log:     logging.NewLogger("Bot"),
		maxWarn: 5,
	}

	bot.initUpdateAdmins()

	return bot
}

func (b *Bot) initUpdateAdmins() {
	go func() {
		b.updateAllAdmins()
		for range time.Tick(time.Hour) {
			b.updateAllAdmins()
		}
	}()
}

// Start : initialize a bot
func (b *Bot) Start() error {
	err := db.Connect(b.log)
	if err != nil {
		return err
	}

	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint
		db.Shutdown()
		b.log.Info("\nBot closed." + time.Now().String())
		os.Exit(0)
	}()

	b.log.Infof("Successfully connected on %s", b.Self.UserName)
	return nil
}

// KeepOn : start messaging and block main function
func (b *Bot) KeepOn() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.GetUpdatesChan(u)
	if err != nil {
		b.log.Fatal("Unable get updates", err)
	}

	for update := range updates {
		msg := update.Message
		if msg != nil {
			go b.processMessage(&model.Message{Message: msg})
		}
	}
}
