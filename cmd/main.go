package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"

	"github.com/defer-panic/news-feed-bot/chat"
	"github.com/defer-panic/news-feed-bot/internal/bot"
	"github.com/defer-panic/news-feed-bot/internal/classifier"
	"github.com/defer-panic/news-feed-bot/internal/fetcher"
	"github.com/defer-panic/news-feed-bot/internal/provider"
	"github.com/defer-panic/news-feed-bot/internal/source"
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

	f, err := os.OpenFile("internal/classifier/data/classes.json", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("[ERROR] failed to open file: %v", err)
		return
	}
	defer f.Close()

	classes, err := classifier.LoadClasses(f)
	if err != nil {
		log.Printf("[ERROR] failed to load classes: %v", err)
		return
	}

	var (
		chatStorage    = storage.NewChatStorage(db)
		articleStorage = storage.NewArticleStorage(db)
		clsfr          = classifier.New(classes)
		manager        = chat.NewManager(chatStorage, clsfr)
		ftchr          = fetcher.New(fetcher.Config{
			Sources:                      prepareSources(),
			FetchInterval:                10 * time.Minute,
			ContentExtractInterval:       10 * time.Second,
			ContentExtractIntervalJitter: 5 * time.Second,
			Classifier:                   clsfr,
			ArticleStorage:               articleStorage,
		})
		p = provider.New(manager, articleStorage, clsfr, botAPI, 15*time.Minute, 1*time.Hour)
	)

	newsBot := bot.New(botAPI)
	newsBot.RegisterCmdView("start", bot.ViewCmdStart(manager))
	newsBot.RegisterCmdView("settopics", bot.ViewCmdSetTopics(clsfr, manager))
	newsBot.RegisterCmdView("latest", bot.ViewCmdLatest(p))
	newsBot.RegisterCallbackView("edit-topics", bot.ViewCallbackEditTopics(manager, manager, clsfr))

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

func prepareSources() []fetcher.Source {
	return []fetcher.Source{
		source.RSSSource{
			SourceName: "Коммерсант",
			URL:        "https://www.kommersant.ru/RSS/news.xml",
		},
		source.RSSSource{
			SourceName: "Хабр",
			URL:        "https://habr.com/ru/rss/news/?fl=ru",
		},
		source.RSSSource{
			SourceName: "Ferra",
			URL:        "https://www.ferra.ru/exports/rss.xml",
		},
		source.NewOsnovaSource("DTF", "https://api.dtf.ru", "<token>"),
		source.NewOsnovaSource("vc.ru", "https://api.vc.ru", "<token>"),
	}
}
