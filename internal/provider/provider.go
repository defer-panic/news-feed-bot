package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type ArticleProvider interface {
	NotPostedArticles(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error)
	MarkArticleAsPosted(ctx context.Context, article model.Article) error
}

type Provider struct {
	articles     ArticleProvider
	bot          *tgbotapi.BotAPI
	sendInterval time.Duration
	chatID       int64
}

func New(
	articleProvider ArticleProvider,
	bot *tgbotapi.BotAPI,
	sendInterval time.Duration,
	chatID int64,
) *Provider {
	return &Provider{
		articles:     articleProvider,
		bot:          bot,
		sendInterval: sendInterval,
		chatID:       chatID,
	}
}

func (p *Provider) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.sendInterval)
	defer ticker.Stop()

	if err := p.SelectAndSendArticle(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := p.SelectAndSendArticle(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Provider) SelectAndSendArticle(ctx context.Context) error {
	topOneArticles, err := p.articles.NotPostedArticles(ctx, time.Now().Add(-20*time.Minute), 1)
	if err != nil {
		return err
	}

	if len(topOneArticles) == 0 {
		return nil
	}

	article := topOneArticles[0]

	if err := p.sendArticle(ctx, article); err != nil {
		return err
	}

	if err := p.articles.MarkArticleAsPosted(ctx, article); err != nil {
		return err
	}

	return nil
}

func (p *Provider) sendArticle(ctx context.Context, article model.Article) error {
	const msgFormat = `[%s](%s)`

	msg := tgbotapi.NewMessage(
		p.chatID,
		fmt.Sprintf(
			msgFormat,
			escapeForMarkdown(article.Title),
			escapeForMarkdown(article.Link),
		),
	)
	msg.ParseMode = "MarkdownV2"

	_, err := p.bot.Send(msg)
	if err != nil {
		return err
	}

	if err := p.articles.MarkArticleAsPosted(ctx, article); err != nil {
		return err
	}

	return nil
}

func escapeForMarkdown(link string) string {
	replacer := strings.NewReplacer(
		"-",
		"\\-",
		"_",
		"\\_",
		"*",
		"\\*",
		"[",
		"\\[",
		"]",
		"\\]",
		"(",
		"\\(",
		")",
		"\\)",
		"~",
		"\\~",
		"`",
		"\\`",
		">",
		"\\>",
		"#",
		"\\#",
		"+",
		"\\+",
		"=",
		"\\=",
		"|",
		"\\|",
		"{",
		"\\{",
		"}",
		"\\}",
		".",
		"\\.",
		"!",
		"\\!",
	)

	return replacer.Replace(link)
}
