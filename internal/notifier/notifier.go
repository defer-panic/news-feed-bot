package notifier

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/defer-panic/news-feed-bot/internal/botkit/markup"
	"github.com/defer-panic/news-feed-bot/internal/model"
)

type ArticleProvider interface {
	AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error)
	MarkAsPosted(ctx context.Context, article model.Article) error
}

type Notifier struct {
	articles         ArticleProvider
	bot              *tgbotapi.BotAPI
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	chatID           int64
}

func New(
	articleProvider ArticleProvider,
	bot *tgbotapi.BotAPI,
	sendInterval time.Duration,
	lookupTimeWindow time.Duration,
	chatID int64,
) *Notifier {
	return &Notifier{
		articles:         articleProvider,
		bot:              bot,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
		chatID:           chatID,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	if err := n.SelectAndSendArticle(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := n.SelectAndSendArticle(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (n *Notifier) SelectAndSendArticle(ctx context.Context) error {
	topOneArticles, err := n.articles.AllNotPosted(ctx, time.Now().Add(-n.lookupTimeWindow), 1)
	if err != nil {
		return err
	}

	if len(topOneArticles) == 0 {
		return nil
	}

	article := topOneArticles[0]

	if err := n.sendArticle(ctx, article); err != nil {
		return err
	}

	if err := n.articles.MarkAsPosted(ctx, article); err != nil {
		return err
	}

	return nil
}

func (n *Notifier) sendArticle(ctx context.Context, article model.Article) error {
	const msgFormat = `[%s](%s)`

	msg := tgbotapi.NewMessage(
		n.chatID,
		fmt.Sprintf(
			msgFormat,
			markup.EscapeForMarkdown(article.Title),
			markup.EscapeForMarkdown(article.Link),
		),
	)
	msg.ParseMode = "MarkdownV2"

	_, err := n.bot.Send(msg)
	if err != nil {
		return err
	}

	if err := n.articles.MarkAsPosted(ctx, article); err != nil {
		return err
	}

	return nil
}
