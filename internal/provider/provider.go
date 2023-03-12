package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type ChatProvider interface {
	ListChats(ctx context.Context) ([]model.Chat, error)
	GetChat(ctx context.Context, telegramID int64) (*model.Chat, error)
	SetTopicsByTelegramID(ctx context.Context, telegramID int64, topics []string) error
}

type ArticleProvider interface {
	StoreArticle(ctx context.Context, article model.ClassifiedItem) error
	GetTopArticles(
		ctx context.Context,
		chatTelegramID int64,
		topics []string,
		n int64,
		timeWindow time.Duration,
	) ([]model.ClassifiedItem, error)
	SaveArticleSent(ctx context.Context, articleID int64, chatTelegramID int64) error
	IsArticleSent(ctx context.Context, articleID int64, chatTelegramID int64) (bool, error)
}

type classProvider interface {
	Classes(ctx context.Context) ([]*model.Class, error)
}

type Provider struct {
	chats         ChatProvider
	articles      ArticleProvider
	classProvider classProvider
	bot           *tgbotapi.BotAPI
	sendInterval  time.Duration
	checkWindow   time.Duration
}

func New(
	chatStorage ChatProvider,
	articleProvider ArticleProvider,
	classProvider classProvider,
	bot *tgbotapi.BotAPI,
	sendInterval time.Duration,
	checkWindow time.Duration,
) *Provider {
	return &Provider{
		chats:         chatStorage,
		articles:      articleProvider,
		classProvider: classProvider,
		bot:           bot,
		sendInterval:  sendInterval,
		checkWindow:   checkWindow,
	}
}

func (p *Provider) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.sendInterval)
	defer ticker.Stop()

	if err := p.SendArticlesToChats(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := p.SendArticlesToChats(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Provider) SendArticlesToChats(ctx context.Context) error {
	chats, err := p.chats.ListChats(ctx)
	if err != nil {
		return err
	}

	for _, chat := range chats {
		topOneArticles, err := p.articles.GetTopArticles(
			ctx,
			chat.TelegramID,
			chat.Topics.Slice(),
			10,
			p.checkWindow,
		)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] topOneArticles for %d: %v", chat.TelegramID, topOneArticles)

		if len(topOneArticles) == 0 {
			continue
		}

		if err := p.sendArticle(context.TODO(), chat, topOneArticles[0]); err != nil {
			log.Printf("[ERROR] failed to send article to chat %d: %v", chat.TelegramID, err)
		}

		time.Sleep(time.Second / 2)
	}

	return nil
}

func (p *Provider) SendLatestToChat(ctx context.Context, chatID int64) error {
	chat, err := p.chats.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	topArticles, err := p.articles.GetTopArticles(ctx, chat.TelegramID, chat.Topics.Slice(), 5, 1*time.Hour)
	if err != nil {
		return err
	}

	if len(topArticles) == 0 {
		if _, err := p.bot.Send(
			tgbotapi.NewMessage(
				chatID,
				"За последний час не было интересных новостей 😢\nПопробуйте выбрать больше тем, написав /settopics",
			),
		); err != nil {
			return err
		}
		return nil
	}

	var msgText strings.Builder

	msgText.WriteString("Последние новости:\n\n")

	for _, article := range topArticles {
		icon, err := p.getClassIcon(ctx, article.Slug)
		if err != nil {
			log.Printf("[ERROR] failed to get class icon: %v", err)
			icon = ""
		}

		msgText.WriteString(fmt.Sprintf(
			"%s %s: [%s](%s)\n\n",
			icon,
			escapeForMarkdown(article.Item.SourceName),
			escapeForMarkdown(article.Item.Title),
			escapeForMarkdown(article.Item.Link),
		))
	}

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ParseMode = "MarkdownV2"

	if _, err := p.bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func (p *Provider) sendArticle(ctx context.Context, chat model.Chat, article model.ClassifiedItem) error {
	const msgFormat = `%s %s: [%s](%s)`

	icon, err := p.getClassIcon(ctx, article.Slug)
	if err != nil {
		log.Printf("[ERROR] failed to get class icon: %v", err)
		icon = ""
	}

	log.Printf(
		"[INFO] sending article (category=%s) to chat %d: %s",
		article.Slug,
		chat.TelegramID,
		article.Item.Title,
	)

	msg := tgbotapi.NewMessage(
		chat.TelegramID,
		fmt.Sprintf(
			msgFormat,
			icon,
			escapeForMarkdown(article.Item.SourceName),
			escapeForMarkdown(article.Item.Title),
			escapeForMarkdown(article.Item.Link),
		),
	)
	msg.ParseMode = "MarkdownV2"

	_, err = p.bot.Send(msg)
	if err != nil {
		return err
	}

	if err := p.articles.SaveArticleSent(ctx, article.ID, chat.TelegramID); err != nil {
		return err
	}

	return nil
}

func (p *Provider) getClassIcon(ctx context.Context, slug string) (string, error) {
	classes, err := p.classProvider.Classes(ctx)
	if err != nil {
		return "", err
	}

	for _, class := range classes {
		if class.Slug == slug {
			return class.Icon, nil
		}
	}

	return "", nil
}

func (p *Provider) formatArticleMsgText(article model.ClassifiedItem) string {
	const msgFormat = `%s: [%s](%s)`

	return fmt.Sprintf(msgFormat, article.Item.SourceName, article.Item.Title, article.Item.Link)
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
