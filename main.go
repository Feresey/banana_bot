package main

import (
	"flag"
	"os"

	"github.com/Feresey/banana_bot/bot"
)

func main() {
	debug := flag.Bool("debug", false, "print all message data")
	flag.Parse()
	token := os.Getenv("TOKEN")
	bot := bot.NewBot(token, *debug)

	if err := bot.Start(); err == nil {
		bot.KeepOn()
	} else {
		panic(err)
	}

}
