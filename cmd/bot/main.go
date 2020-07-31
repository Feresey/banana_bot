package main

import (
	"flag"
	"os"

	"github.com/Feresey/banana_bot/bot"
	"github.com/spf13/viper"
)

func main() {
	configPath := flag.String("c", "", "config path")
	flag.Parse()
	token := os.Getenv("TOKEN")

	var config bot.Config

	v := viper.New()
	v.SetConfigFile(*configPath)
	v.SetDefault("token", token)
	if err := v.Unmarshal(&config); err != nil {
		panic(err)
	}

	b := bot.New(config)
	b.Init()
	b.Start()
}
