package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Feresey/banana_bot/internal/db"
	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Token         string
	MaxConcurrent int
	MaxWarn       int64
	DBConfig      db.Config

	ApiTimeout    time.Duration
	ResponseSleep time.Duration

	LogFile string
}

//go:generate go run github.com/golang/mock/mockgen -destination mock_test.go -package bot . TelegramAPI,Database

type TelegramAPI interface {
	GetUpdatesChan(tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error)
	StopReceivingUpdates()
	KickChatMember(tgbotapi.KickChatMemberConfig) (*tgbotapi.APIResponse, error)
	Send(tgbotapi.Chattable) (*tgbotapi.Message, error)
	DeleteMessage(tgbotapi.DeleteMessageConfig) (*tgbotapi.APIResponse, error)
	GetChatMember(tgbotapi.ChatConfigWithUser) (*tgbotapi.ChatMember, error)
	RestrictChatMember(tgbotapi.RestrictChatMemberConfig) (*tgbotapi.APIResponse, error)
	IsMessageToMe(*tgbotapi.Message) bool
	GetChat(config tgbotapi.ChatConfig) (*tgbotapi.Chat, error)
	AnswerCallbackQuery(tgbotapi.CallbackConfig) (*tgbotapi.APIResponse, error)
}

type Database interface {
	Init(context.Context) error
	Shutdown(context.Context) error
	Warn(ctx context.Context, person *db.Person, add bool) (int64, error)
	Subscribe(ctx context.Context, p *db.Person) error
	Unsubscribe(ctx context.Context, p *db.Person) error
	Report(ctx context.Context, chatID int64) (res []int64, err error)

	GetMyChats(ctx context.Context) ([]int64, error)
	AddChatWithMe(ctx context.Context, chatID int64) error
	DelChatWithMe(ctx context.Context, chatID int64) error
}

type Bot struct {
	c   *Config
	log *zap.Logger

	done    chan struct{}
	updates <-chan tgbotapi.Update
	api     TelegramAPI
	me      *tgbotapi.User

	cs    *callbackStorage
	chats sync.Map

	db Database
}

func New(config Config) *Bot {
	return &Bot{
		c:    &config,
		done: make(chan struct{}),
	}
}

func (b *Bot) Init() {
	console := zap.NewDevelopmentEncoderConfig()
	console.EncodeLevel = zapcore.CapitalColorLevelEncoder
	console.EncodeCaller = zapcore.FullCallerEncoder
	consoleOut, _, err := zap.Open("stdout")
	if err != nil {
		log.Fatal(err)
	}

	file := zap.NewProductionEncoderConfig()
	file.EncodeCaller = zapcore.ShortCallerEncoder
	fileOut, _, err := zap.Open(b.c.LogFile)
	if err != nil {
		log.Fatal(err)
	}

	log := zap.New(
		zapcore.NewTee(
			zapcore.NewCore(zapcore.NewConsoleEncoder(console), consoleOut, zapcore.DebugLevel),
			zapcore.NewCore(zapcore.NewJSONEncoder(file), fileOut, zapcore.DebugLevel),
		),
		zap.AddCaller(),
		zap.AddStacktrace(zap.FatalLevel),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddStacktrace(zap.WarnLevel),
	)

	b.log = log.Named("bot")
	log = log.Named("init")
	log.Info("Init")

	b.cs = NewCallbackStorage(b.log)

	err = b.initAPI(log)
	if err != nil {
		log.Fatal("Connect to telegram api", zap.Error(err))
	}

	err = b.initDB(log)
	if err != nil {
		log.Fatal("Connect to DB", zap.Error(err))
	}

	err = b.initUpdates(log)
	if err != nil {
		log.Fatal("Get updates", zap.Error(err))
	}

	log.Info("Bot ready")
}

func (b *Bot) initAPI(log *zap.Logger) error {
	cli := &http.Client{
		Timeout:   b.c.ApiTimeout,
		Transport: &http.Transport{},
	}

	api, err := tgbotapi.NewBotAPIWithClient(b.c.Token, tgbotapi.APIEndpoint, cli)
	if err != nil {
		return fmt.Errorf("connect to telegram api: %w", err)
	}
	b.api = api
	b.me, err = api.GetMe()
	if err != nil {
		return fmt.Errorf("get me: %w", err)
	}
	log.Info("Connected to api", zap.String("username", b.me.String()))
	return nil
}

func (b *Bot) initDB(log *zap.Logger) error {
	log.Info("Init DB", zap.Duration("connect_timeout", b.c.DBConfig.ConnectTimeout))

	b.db = db.New(b.log, b.c.DBConfig)
	ctx, cancel := context.WithTimeout(context.Background(), b.c.DBConfig.ConnectTimeout)
	defer cancel()

	err := b.db.Init(ctx)
	if err != nil {
		return err
	}
	log.Info("Database connected")
	return nil
}

func (b *Bot) initUpdates(log *zap.Logger) error {
	// переподписка?
	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(b.c.ApiTimeout.Seconds())
	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("get updates %w", err)
	}
	b.updates = updates
	log.Info("Subscribe on updates")
	return nil
}

func (b *Bot) Start() {
	// тут немного костылики, т.к. канал завершения лежит в структуре бота.
	// Это сделано ради того, чтобы не передавать канал в bot.Shutdown
	b.Listen()

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
	if b.api != nil {
		_ = b.pushLogs()
		b.log.Info("Stopping updates")
		b.api.StopReceivingUpdates()
		// fuck this
		// <-b.done
		b.log.Info("Updates stopped")
	}
	if b.db != nil {
		if err := b.db.Shutdown(ctx); err != nil {
			multi = multierr.Append(multi, err)
		}
	}
	return multi
}

// Listen открывает стрим сообщений (обновлений) от телеги. Если возникает ошибка,
// то библиотека срёт в лог и закрывает канал. Я ебал, зачем так делать?
func (b *Bot) Listen() {
	go func() {
		b.listen()
		close(b.done)
	}()
}

func (b *Bot) listen() {
	defer b.log.Info("Listening cancelled")

	limit := make(chan struct{}, b.c.MaxConcurrent)
	for update := range b.updates {
		if update.CallbackQuery != nil {
			b.cs.HandleCallabck(update.CallbackQuery)
			continue
		}

		msg := update.Message
		if update.EditedMessage != nil {
			msg = update.EditedMessage
		}
		if msg == nil {
			b.log.Debug("Non-message update", zap.Reflect("data", update))
			continue
		}
		b.log.Info("Got a message")

		limit <- struct{}{}

		go b.addChatIfNeeded(msg.Chat)

		go func() {
			err := b.processMessage(msg)
			if err != nil {
				b.log.Error("Process message failed", zap.Error(err))
			}
			<-limit
		}()
	}
}

func (b *Bot) addChatIfNeeded(chat *tgbotapi.Chat) {
	if !(chat.IsGroup() || chat.IsSuperGroup()) {
		return
	}

	_, ok := b.chats.LoadOrStore(chat.ID, nil)
	if ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := b.db.AddChatWithMe(ctx, chat.ID); err != nil {
		b.log.Warn("Save new chat",
			zap.Error(err),
			zap.Int64("chat_id", chat.ID),
			zap.String("chat_name", chat.Title),
		)
	}
}
