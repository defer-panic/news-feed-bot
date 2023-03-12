package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/samber/lo"
	"github.com/tomakado/containers/set"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type ChatPostgresStorage struct {
	db *sqlx.DB
}

func NewChatStorage(db *sqlx.DB) *ChatPostgresStorage {
	return &ChatPostgresStorage{db: db}
}

func (s *ChatPostgresStorage) StoreChat(ctx context.Context, chat model.Chat) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(
		ctx,
		`INSERT INTO chats (telegram_id, topics ) 
				VALUES ($1, $2)
				ON CONFLICT (telegram_id) DO UPDATE SET topics = $2, updated_at = now()`,
		chat.TelegramID,
		pq.StringArray(chat.Topics.Slice()),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *ChatPostgresStorage) GetChatByTelegramID(ctx context.Context, telegramID int64) (*model.Chat, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var chat dbChat
	if err := conn.GetContext(ctx, &chat, `SELECT * FROM chats WHERE telegram_id = $1 LIMIT 1;`, telegramID); err != nil {
		return nil, err
	}

	mappedChat := mapDbChatToChat(chat)

	return &mappedChat, nil
}

func (s *ChatPostgresStorage) ListChats(ctx context.Context) ([]model.Chat, error) {
	// TODO: add pagination

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var dbChats []dbChat

	if err := conn.SelectContext(ctx, &dbChats, "SELECT * FROM chats"); err != nil {
		return nil, err
	}

	return lo.Map(dbChats, func(chat dbChat, _ int) model.Chat {
		return mapDbChatToChat(chat)
	}), nil
}

func (s *ChatPostgresStorage) SetTopicsByTelegramID(ctx context.Context, telegramID int64, topics []string) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(
		ctx, `UPDATE chats 
					SET topics = $1, updated_at = now() 
					WHERE telegram_id = $3`,
		pq.StringArray(topics), time.Now(), telegramID)
	if err != nil {
		return err
	}

	return nil
}

type dbChat struct {
	ID         int64          `db:"id"`
	TelegramID int64          `db:"telegram_id"`
	Topics     pq.StringArray `db:"topics"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  sql.NullTime   `db:"updated_at"`
}

func mapDbChatToChat(chat dbChat) model.Chat {
	return model.Chat{
		ID:         chat.ID,
		TelegramID: chat.TelegramID,
		Topics:     set.New[string](chat.Topics...),
		CreatedAt:  chat.CreatedAt,
		UpdatedAt:  chat.UpdatedAt.Time,
	}
}
