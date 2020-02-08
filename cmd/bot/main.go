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

	if err := bot.Start(token, *debug); err == nil {
		bot.KeepOn()
	} else {
		panic(err)
	}
}
