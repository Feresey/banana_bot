package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Feresey/banana_bot/db"
	"github.com/Feresey/banana_bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

var defaultConfig = Config{
	Token:         "", // strong protection!
	MaxConcurrent: 10,
	MaxWarn:       5,
}

type Config struct {
	Token         string
	MaxConcurrent int
	MaxWarn       int
	DBConfig      db.Config
}

type Bot struct {
	c *Config

	log *zap.Logger

	api *tgbotapi.BotAPI
	db  *db.Database

	updates <-chan tgbotapi.Update
	done    chan struct{}
}

func New(config Config) *Bot {
	return &Bot{
		c: &config,

		done: make(chan struct{}),
	}
}

func (b *Bot) Init() error {
	log, err := zap.NewProduction()
	if err != nil {
		return err
	}
	b.log = log

	api, err := tgbotapi.NewBotAPI(b.c.Token)
	if err != nil {
		return fmt.Errorf("connect to telegram api: %w", err)
	}
	b.api = api
	log.Info("Connected to api", zap.String("username", b.api.Self.UserName))

	b.db = db.New(log, b.c.DBConfig)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = b.db.Init(ctx)
	if err != nil {
		return err
	}
	log.Info("Connected to DB")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("get updates: %w", err)
	}
	b.updates = updates
	log.Info("Subscribe on updates")

	return nil
}

func (b *Bot) Start() {
	go func() {
		b.listen()
		close(b.done)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	s := <-sig
	b.log.Info("Signal received", zap.Stringer("signal", s))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := b.Shutdown(ctx); err != nil {
		b.log.Error("Shutdown", zap.Error(err))
		os.Exit(1)
	}

	b.log.Info("Bot stopped")
	os.Exit(0)
}

func (b *Bot) Shutdown(ctx context.Context) (multi error) {
	b.api.StopReceivingUpdates()
	<-b.done
	if err := b.db.Shutdown(ctx); err != nil {
		multi = multierror.Append(multi, err)
	}
	return multi
}

// KeepOn : start messaging and block main function
func (b *Bot) listen() {
	defer b.log.Info("Listening cancelled")

	limit := make(chan struct{}, b.c.MaxConcurrent)
	for update := range b.updates {
		msg := update.Message
		if msg == nil {
			data, _ := json.MarshalIndent(update, "", "  ")
			b.log.Debug("Non-message update", zap.ByteString("data", data))
			continue
		}

		limit <- struct{}{}
		go func() {
			processMessage(model.Message{Message: msg})
			<-limit
		}()
	}
}
