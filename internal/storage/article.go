package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type ArticlePostgresStorage struct {
	db *sqlx.DB
}

func NewArticleStorage(db *sqlx.DB) *ArticlePostgresStorage {
	return &ArticlePostgresStorage{db: db}
}

func (s *ArticlePostgresStorage) StoreArticle(ctx context.Context, article model.ClassifiedItem) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`INSERT INTO articles (title, link, topic, topic_score, source_name, published_at)
	    				VALUES ($1, $2, $3, $4, $5, $6)
	    				ON CONFLICT DO NOTHING ;`,
		article.Item.Title,
		article.Item.Link,
		article.Slug,
		article.Score,
		article.Item.SourceName,
		article.Item.Date.UTC(),
	); err != nil {
		return err
	}

	return nil
}

func (s *ArticlePostgresStorage) GetTopArticles(
	ctx context.Context,
	chatTelegramID int64,
	topics []string,
	n int64,
	timeWindow time.Duration,
) ([]model.ClassifiedItem, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var articles []dbArticle
	if err := conn.SelectContext(
		ctx,
		&articles,
		`SELECT * FROM articles 
         		WHERE topic IN (SELECT * FROM unnest($2::text[]))
         		AND published_at >= $3::timestamp
         		AND (SELECT COUNT(*) FROM sent_articles WHERE article_id = articles.id AND chat_telegram_id = $1) = 0
         		ORDER BY topic_score DESC
         		LIMIT $4;
			`,
		chatTelegramID,
		pq.StringArray(topics),
		time.Now().UTC().Add(-timeWindow).Format(time.RFC3339),
		n,
	); err != nil {
		return nil, err
	}

	classifiedItems := make([]model.ClassifiedItem, 0, len(articles))
	for _, article := range articles {
		classifiedItems = append(classifiedItems, model.ClassifiedItem{
			ID:    article.ID,
			Slug:  article.Topic,
			Score: article.TopicScore,
			Item: model.Item{
				Title:      article.Title,
				Link:       article.Link,
				Date:       article.PublishedAt,
				SourceName: article.SourceName,
			},
		})
	}

	return classifiedItems, nil
}

func (s *ArticlePostgresStorage) SaveArticleSent(ctx context.Context, articleID int64, chatTelegramID int64) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`INSERT INTO sent_articles (article_id, chat_telegram_id)
	    				VALUES ($1, $2)
	    				ON CONFLICT DO NOTHING ;`,
		articleID,
		chatTelegramID,
	); err != nil {
		return err
	}

	return nil
}

func (s *ArticlePostgresStorage) IsArticleSent(ctx context.Context, articleID int64, chatTelegramID int64) (bool, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	var count int
	if err := conn.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*) FROM sent_articles 
		 		WHERE article_id = $1
		 		AND chat_telegram_id = $2;`,
		articleID,
		chatTelegramID,
	); err != nil {
		return false, err
	}

	return count > 0, nil
}

type dbArticle struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Link        string    `db:"link"`
	Topic       string    `db:"topic"`
	TopicScore  int       `db:"topic_score"`
	SourceName  string    `db:"source_name"`
	PublishedAt time.Time `db:"published_at"`
	CreatedAt   time.Time `db:"created_at"`
}
