package config

import (
	"log"
	"sync"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"
)

type Config struct {
	TelegramBotToken     string        `hcl:"telegram_bot_token"`
	TelegramChannelID    int64         `hcl:"telegram_channel_id"`
	DatabaseDSN          string        `hcl:"database_dsn" default:"postgres://postgres:postgres@localhost:5432/news_feed_bot?sslmode=disable"`
	FetchInterval        time.Duration `hcl:"fetch_interval" default:"10m"`
	NotificationInterval time.Duration `hcl:"notification_interval" default:"1m"`
}

var (
	cfg  Config
	once sync.Once
)

func Get() Config {
	once.Do(func() {
		loader := aconfig.LoaderFor(&cfg, aconfig.Config{
			EnvPrefix: "NFB",
			Files:     []string{"./config.hcl", "./config.local.hcl", "$HOME/.config/news-feed-bot/config.hcl"},
			FileDecoders: map[string]aconfig.FileDecoder{
				".hcl": aconfighcl.New(),
			},
		})

		if err := loader.Load(); err != nil {
			log.Printf("[ERROR] failed to load config: %v", err)
		}
	})

	return cfg
}
