package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/defer-panic/news-feed-bot/internal/botkit"
	"github.com/defer-panic/news-feed-bot/internal/model"
)

type SourceStorage interface {
	Add(ctx context.Context, source model.Source) (int64, error)
}

func ViewCmdAddSource(storage SourceStorage) botkit.ViewFunc {
	type addSourceArgs struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Priority int    `json:"priority"`
	}

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[addSourceArgs](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		source := model.Source{
			Name:     args.Name,
			FeedURL:  args.URL,
			Priority: args.Priority,
		}

		sourceID, err := storage.Add(ctx, source)
		if err != nil {
			// TODO: send error message
			return err
		}

		var (
			msgText = fmt.Sprintf(
				"Источник добавлен с ID: `%d`\\. Используйте этот ID для обновления источника или удаления\\.",
				sourceID,
			)
			reply = tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		)

		reply.ParseMode = "MarkdownV2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
