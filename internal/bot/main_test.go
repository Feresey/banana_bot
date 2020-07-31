package bot

import (
	"os"
	"testing"
	"time"

	"github.com/Feresey/banana_bot/internal/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var botConfig = &Config{
	Token:         "token",
	ApiTimeout:    5 * time.Second,
	ResponseSleep: time.Millisecond,
	DBConfig: db.Config{
		LogLevel:       "debug",
		ConnectTimeout: 5 * time.Second,
		Migrate:        10,
		SQL:            "sql url",
	},
	MaxConcurrent: 10,
	MaxWarn:       5,
}

var (
	bot     *Bot
	updates = make(chan tgbotapi.Update)
)

func TestMain(m *testing.M) {
	lc := zap.NewDevelopmentConfig()
	lc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	log, err := lc.Build()
	if err != nil {
		panic(err)
	}
	bot = &Bot{
		c: botConfig,

		done: make(chan struct{}),
		log:  log.Named("test"),

		updates: updates,
	}
	os.Exit(m.Run())
}
