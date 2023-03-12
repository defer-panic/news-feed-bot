package chat

import (
	"context"

	"github.com/samber/lo"
	"github.com/tomakado/containers/set"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type Storage interface {
	StoreChat(ctx context.Context, chat model.Chat) error
	GetChatByTelegramID(ctx context.Context, telegramID int64) (*model.Chat, error)
	ListChats(ctx context.Context) ([]model.Chat, error)
	SetTopicsByTelegramID(ctx context.Context, telegramID int64, topics []string) error
}

type ClassProvider interface {
	Classes(ctx context.Context) ([]*model.Class, error)
}

type Manager struct {
	chats         Storage
	classProvider ClassProvider
}

func NewManager(chats Storage, classProvider ClassProvider) *Manager {
	return &Manager{
		chats:         chats,
		classProvider: classProvider,
	}
}

func (m *Manager) RegisterChat(ctx context.Context, chatTelegramID int64) error {
	classes, err := m.classProvider.Classes(ctx)
	if err != nil {
		return err
	}

	classSlugs := lo.Map(classes, func(class *model.Class, _ int) string {
		return class.Slug
	})

	return m.chats.StoreChat(ctx, model.Chat{
		TelegramID: chatTelegramID,
		Topics:     set.New[string](classSlugs...),
	})
}

func (m *Manager) GetChat(ctx context.Context, telegramID int64) (*model.Chat, error) {
	return m.chats.GetChatByTelegramID(ctx, telegramID)
}

func (m *Manager) ListChats(ctx context.Context) ([]model.Chat, error) {
	return m.chats.ListChats(ctx)
}

func (m *Manager) AddTopic(ctx context.Context, telegramID int64, topic string) error {
	chat, err := m.chats.GetChatByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	chat.Topics.Add(topic)

	return m.chats.StoreChat(ctx, *chat)
}

func (m *Manager) RemoveTopic(ctx context.Context, telegramID int64, topic string) error {
	chat, err := m.chats.GetChatByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	chat.Topics.Remove(topic)

	return m.chats.StoreChat(ctx, *chat)
}

func (m *Manager) SetTopicsByTelegramID(ctx context.Context, telegramID int64, topics []string) error {
	return m.chats.SetTopicsByTelegramID(ctx, telegramID, topics)
}
