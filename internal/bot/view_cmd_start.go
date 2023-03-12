package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type chatRegistrar interface {
	RegisterChat(ctx context.Context, chatTelegramID int64) error
}

func ViewCmdStart(registrar chatRegistrar) ViewFunc {
	return func(ctx context.Context, botAPI *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if err := registrar.RegisterChat(ctx, update.Message.Chat.ID); err != nil {
			return err
		}

		const format = `Привет, %s! Я буду присылать вам овости и интересные статьи из разных источников. Чтобы задать интерсующие тебя темы, введите /settopics`

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(format, update.Message.From.FirstName))

		if _, err := botAPI.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
