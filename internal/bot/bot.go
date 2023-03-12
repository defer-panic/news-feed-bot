package bot

import (
	"context"
	"encoding/json"
	"log"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api           *tgbotapi.BotAPI
	cmdViews      map[string]ViewFunc
	callbackViews map[string]ViewFunc
}

func New(api *tgbotapi.BotAPI) *Bot {
	return &Bot{api: api}
}

func (b *Bot) RegisterCmdView(cmd string, view ViewFunc) {
	if b.cmdViews == nil {
		b.cmdViews = make(map[string]ViewFunc)
	}

	b.cmdViews[cmd] = view
}

func (b *Bot) RegisterCallbackView(procedure string, view ViewFunc) {
	if b.callbackViews == nil {
		b.callbackViews = make(map[string]ViewFunc)
	}

	b.callbackViews[procedure] = view
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			b.handleUpdate(updateCtx, update)
			updateCancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("[ERROR] panic recovered: %v\n%s", p, string(debug.Stack()))
		}
	}()

	if (update.Message == nil || !update.Message.IsCommand()) && update.CallbackQuery == nil {
		return
	}

	var view ViewFunc

	switch {
	case update.Message != nil:
		if !update.Message.IsCommand() {
			break
		}

		cmd := update.Message.Command()
		cmdView, ok := b.cmdViews[cmd]
		if !ok {
			return
		}
		view = cmdView
	case update.CallbackQuery != nil:
		var callbackRequest CallbackRequest

		if err := json.Unmarshal([]byte(update.CallbackQuery.Data), &callbackRequest); err != nil {
			log.Printf("[ERROR] failed to parse callback request: %v", err)
			return
		}

		callbackView, ok := b.callbackViews[callbackRequest.Procedure]
		if !ok {
			return
		}

		view = callbackView
	default:
		return
	}

	if err := view(ctx, b.api, update); err != nil {
		log.Printf("[ERROR] failed to execute view: %v", err)

		if _, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error")); err != nil {
			log.Printf("[ERROR] failed to send error message: %v", err)
		}
	}
}

type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, msg tgbotapi.Update) error
