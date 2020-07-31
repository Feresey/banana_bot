package main

import (
	"flag"
	"os"
	"time"

	"github.com/Feresey/banana_bot/internal/bot"
	"github.com/Feresey/banana_bot/internal/db"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

func main() {
	configPath := flag.String("c", "", "config path")
	flag.Parse()
	token := os.Getenv("TOKEN")

	config := bot.Config{
		Token:         token,
		MaxConcurrent: 10,
		MaxWarn:       5,

		ApiTimeout:    time.Minute,
		ResponseSleep: time.Second,

		DBConfig: db.Config{
			// fucking legacy. What about `iota`?!?!?
			LogLevel:       pgx.LogLevel(pgx.LogLevelDebug).String(),
			ConnectTimeout: 10 * time.Second,
			// do migrate
			Migrate: 0,
			SQL:     "postgres://postgres:5432",
		},
	}

	v := viper.New()
	v.SetConfigFile(*configPath)
	if err := v.Unmarshal(&config); err != nil {
		panic(err)
	}

	b := bot.New(config)
	b.Init()
	b.Start()
}
