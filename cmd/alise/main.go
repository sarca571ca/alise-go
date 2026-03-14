package main

import (
	"alise-go/internal/bot"
	"alise-go/internal/config"
	"context"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := b.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
