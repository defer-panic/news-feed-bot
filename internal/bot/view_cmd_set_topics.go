package bot

import (
	"context"
	"fmt"
	"log"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type classProvider interface {
	Classes(ctx context.Context) ([]*model.Class, error)
}

type chatManager interface {
	GetChat(ctx context.Context, telegramID int64) (*model.Chat, error)
	SetTopicsByTelegramID(ctx context.Context, telegramID int64, topics []string) error
}

func ViewCmdSetTopics(classProvider classProvider, chatManager chatManager) ViewFunc {
	return func(ctx context.Context, botAPI *tgbotapi.BotAPI, update tgbotapi.Update) error {
		msg, err := buildTopicsMessage(ctx, update, classProvider, chatManager)
		if err != nil {
			return err
		}

		if _, err := botAPI.Send(msg); err != nil {
			return err
		}

		return nil
	}
}

func buildTopicsMessage(
	ctx context.Context,
	update tgbotapi.Update,
	classProvider classProvider,
	chatManager chatManager,
) (tgbotapi.MessageConfig, error) {
	classes, err := classProvider.Classes(ctx)
	if err != nil {
		log.Printf("[ERROR] failed to get classes: %v", err)
		return tgbotapi.MessageConfig{}, nil
	}

	sort.SliceStable(classes, func(i, j int) bool {
		return classes[i].Name < classes[j].Name
	})

	chat, err := chatManager.GetChat(ctx, update.FromChat().ID)
	if err != nil {
		log.Printf("[ERROR] failed to get chat: %v", err)
		return tgbotapi.MessageConfig{}, nil
	}

	const format = "Выберите интересующие вас темы.\n\n[✔] — тема выбрана"

	var (
		buttons = lo.Map(classes, func(class *model.Class, _ int) []tgbotapi.InlineKeyboardButton {
			var (
				prefix          string
				callbackRequest = topicChangeRequest{
					Topic:  class.Slug,
					Action: topicActionAdd,
				}
			)

			if chat.Topics.Contains(class.Slug) {
				prefix = "[✔]"
				callbackRequest.Action = topicActionRemove
			}

			callbackDataJSON, err := MarshalCallbackRequest("edit-topics", callbackRequest)
			if err != nil {
				log.Printf("[ERROR] failed to render inline button: %v", err)
				return nil
			}
			callbackDataStr := string(callbackDataJSON)

			return []tgbotapi.InlineKeyboardButton{
				{
					Text:         fmt.Sprintf("%s %s", prefix, class.Name),
					CallbackData: &callbackDataStr,
				},
			}
		})
		markup  = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
		message = tgbotapi.NewMessage(update.FromChat().ID, format)
	)

	message.ReplyMarkup = markup

	return message, nil
}

type topicChangeRequest struct {
	Topic  string      `json:"t"`
	Action topicAction `json:"a"`
}

type topicAction uint8

const (
	topicActionAdd topicAction = iota + 1
	topicActionRemove
)
