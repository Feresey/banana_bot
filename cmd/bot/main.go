package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Feresey/banana_bot/internal/bot"
	"github.com/Feresey/banana_bot/internal/db"
	"github.com/jackc/pgx/v4"
	"gopkg.in/yaml.v3"
)

var defaultConfig = bot.Config{
	MaxConcurrent: 10,
	MaxWarn:       5,

	ApiTimeout:    10 * time.Second,
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

func readConfig(path string) bot.Config {
	var config bot.Config

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print("Error reading config, fallback to default. ", err)
		return defaultConfig
	}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		log.Print("Error reading config, fallback to default. ", err)
		return defaultConfig
	}
	return config
}

func main() {
	configPath := flag.String("c", "", "config path")
	flag.Parse()

	config := readConfig(*configPath)
	token := os.Getenv("TOKEN")
	if token != "" {
		config.Token = token
	}

	b := bot.New(config)
	b.Init()
	b.Start()
}
