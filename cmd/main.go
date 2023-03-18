package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/defer-panic/news-feed-bot/internal/bot"
	"github.com/defer-panic/news-feed-bot/internal/config"
	"github.com/defer-panic/news-feed-bot/internal/fetcher"
	"github.com/defer-panic/news-feed-bot/internal/notifier"
	"github.com/defer-panic/news-feed-bot/internal/storage"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("[ERROR] failed to create botAPI: %v", err)
		return
	}

	botAPI.Debug = true

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("[ERROR] failed to connect to db: %v", err)
		return
	}
	defer db.Close()

	var (
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage  = storage.NewSourceStorage(db)
		fetcher        = fetcher.New(articleStorage, sourceStorage, config.Get().FetchInterval)
		notifier       = notifier.New(
			articleStorage,
			botAPI,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
			config.Get().TelegramChannelID,
		)
	)

	newsBot := bot.New(botAPI)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			log.Printf("[ERROR] failed to run fetcher: %v", err)
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := notifier.Start(ctx); err != nil {
			log.Printf("[ERROR] failed to start notifier: %v", err)
		}
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		log.Printf("[ERROR] failed to run bot: %v", err)
	}
}
