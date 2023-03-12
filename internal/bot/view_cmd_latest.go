package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type provider interface {
	SendLatestToChat(ctx context.Context, chatID int64) error
}

func ViewCmdLatest(provider provider) ViewFunc {
	return func(ctx context.Context, botAPI *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if err := provider.SendLatestToChat(ctx, update.FromChat().ID); err != nil {
			return err
		}

		return nil
	}
}
