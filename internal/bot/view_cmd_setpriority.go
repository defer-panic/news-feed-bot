package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/defer-panic/news-feed-bot/internal/botkit"
)

type PrioritySetter interface {
	SetPriority(ctx context.Context, sourceID int64, priority int) error
}

func ViewCmdSetPriority(prioritySetter PrioritySetter) botkit.ViewFunc {
	type setPriorityArgs struct {
		SourceID int64 `json:"source_id"`
		Priority int   `json:"priority"`
	}

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[setPriorityArgs](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		if err := prioritySetter.SetPriority(ctx, args.SourceID, args.Priority); err != nil {
			return err
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Приоритет успешно обновлен")

		if _, err := bot.Send(msg); err != nil {
			return err
		}

		return nil
	}
}
