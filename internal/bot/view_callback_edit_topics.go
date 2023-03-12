package bot

import (
	"context"
	"encoding/json"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type topicEditor interface {
	AddTopic(ctx context.Context, chatID int64, topic string) error
	RemoveTopic(ctx context.Context, chatID int64, topic string) error
}

func ViewCallbackEditTopics(topicEditor topicEditor, chatManager chatManager, classProvider classProvider) ViewFunc {
	return func(ctx context.Context, botAPI *tgbotapi.BotAPI, update tgbotapi.Update) error {
		var (
			chatID = update.CallbackQuery.Message.Chat.ID
			data   = json.RawMessage(update.CallbackQuery.Data)
		)

		request, err := UnmarshalCallbackRequest[topicChangeRequest](data)
		if err != nil {
			return err
		}

		var changeFunc func(ctx context.Context, chatID int64, topic string) error

		switch request.Action {
		case topicActionAdd:
			changeFunc = topicEditor.AddTopic
		case topicActionRemove:
			changeFunc = topicEditor.RemoveTopic
		default:
			return errors.New("unknown action")
		}

		if err := changeFunc(ctx, chatID, request.Topic); err != nil {
			return err
		}

		updatedMsg, err := buildTopicsMessage(ctx, update, classProvider, chatManager)
		if err != nil {
			return err
		}

		edit := tgbotapi.NewEditMessageReplyMarkup(
			chatID,
			update.CallbackQuery.Message.MessageID,
			updatedMsg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup),
		)

		if _, err := botAPI.Send(edit); err != nil {
			return err
		}

		return nil
	}
}
