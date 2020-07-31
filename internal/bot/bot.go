package bot

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Feresey/banana_bot/internal/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

type Config struct {
	Token         string
	MaxConcurrent int
	MaxWarn       int64
	DBConfig      db.Config

	ApiTimeout    time.Duration
	ResponseSleep time.Duration
}

//go:generate go run github.com/golang/mock/mockgen -destination mock.go -package bot . TelegramAPI

type TelegramAPI interface {
	GetUpdatesChan(tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error)
	StopReceivingUpdates()
	KickChatMember(tgbotapi.KickChatMemberConfig) (tgbotapi.APIResponse, error)
	Send(tgbotapi.Chattable) (tgbotapi.Message, error)
	DeleteMessage(tgbotapi.DeleteMessageConfig) (tgbotapi.APIResponse, error)
	GetChatMember(tgbotapi.ChatConfigWithUser) (tgbotapi.ChatMember, error)
	RestrictChatMember(tgbotapi.RestrictChatMemberConfig) (tgbotapi.APIResponse, error)
}

type Bot struct {
	c *Config

	log *zap.Logger

	api TelegramAPI
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

func (b *Bot) Init() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	b.log = log

	api, err := tgbotapi.NewBotAPI(b.c.Token)
	if err != nil {
		log.Fatal("Connect to telegram api", zap.Error(err))
	}
	b.api = api
	me, err := api.GetMe()
	if err != nil {
		log.Fatal("GetMe", zap.Error(err))
	}
	log.Info("Connected to api", zap.String("username", me.String()))

	err = b.initDB(log)
	if err != nil {
		log.Fatal("Connect to DB", zap.Error(err))
	}
	log.Info("Connected to DB")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Get updates", zap.Error(err))
	}
	b.updates = updates
	log.Info("Subscribe on updates")

	log.Info("Bot ready")
}

func (b *Bot) initDB(log *zap.Logger) error {
	log = log.Named("db-connect").With(zap.Duration("connect_timeout", b.c.DBConfig.ConnectTimeout))
	log.Info("Init DB")

	b.db = db.New(log, b.c.DBConfig)
	ctx, cancel := context.WithTimeout(context.Background(), b.c.DBConfig.ConnectTimeout)
	defer cancel()

	return b.db.Init(ctx)
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
		b.log.Fatal("Shutdown", zap.Error(err))
	}

	b.log.Info("Bot stopped gracefully")
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
			err := b.processMessage(*msg)
			if err != nil {
				b.log.Error("Process message failed", zap.Error(err))
			}
			<-limit
		}()
	}
}
