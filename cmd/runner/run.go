package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/gehirndienst/supernova-go-bot/internal/botapi"
)

func main() {
	envFile := flag.String("env-file", "../../.env", "Path to .env file from the pwd(!)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	bot, err := botapi.InitBot(*envFile)
	if err != nil {
		panic(err)
	}

	bot.Run(ctx)
}
