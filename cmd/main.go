package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"

	"github.com/defer-panic/news-feed-bot/internal/bot"
	"github.com/defer-panic/news-feed-bot/internal/fetcher"
	"github.com/defer-panic/news-feed-bot/internal/provider"
	"github.com/defer-panic/news-feed-bot/internal/storage"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI("<token>")
	if err != nil {
		log.Printf("[ERROR] failed to create botAPI: %v", err)
		return
	}

	botAPI.Debug = true

	db, err := sqlx.Connect(
		"postgres",
		"postgres://postgres:postgres@localhost:5432/news_feed_bot?sslmode=disable",
	)
	if err != nil {
		log.Printf("[ERROR] failed to connect to db: %v", err)
		return
	}
	defer db.Close()

	var (
		chatID         = int64(1)
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage  = storage.NewSourceStorage(db)
		ftchr          = fetcher.New(fetcher.Config{
			SourcesProvider: sourceStorage,
			FetchInterval:   10 * time.Minute,
			ArticleStorage:  articleStorage,
		})
		p = provider.New(articleStorage, botAPI, 1*time.Minute, chatID)
	)

	newsBot := bot.New(botAPI)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func(ctx context.Context) {
		if err := ftchr.Start(ctx); err != nil {
			log.Printf("[ERROR] failed to run fetcher: %v", err)
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := p.Start(ctx); err != nil {
			log.Printf("[ERROR] failed to start provider: %v", err)
		}
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		log.Printf("[ERROR] failed to run bot: %v", err)
	}
}
