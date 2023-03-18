package fetcher

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/defer-panic/news-feed-bot/internal/model"
	src "github.com/defer-panic/news-feed-bot/internal/source"
)

type ArticleStorage interface {
	Store(ctx context.Context, article model.Article) error
}

type SourcesProvider interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Fetcher struct {
	articles ArticleStorage
	sources  SourcesProvider

	fetchInterval time.Duration
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Config struct {
	FetchInterval   time.Duration
	ArticleStorage  ArticleStorage
	SourcesProvider SourcesProvider
}

func New(articleStorage ArticleStorage, sourcesProvider SourcesProvider, fetchInterval time.Duration) *Fetcher {
	return &Fetcher{
		articles:      articleStorage,
		sources:       sourcesProvider,
		fetchInterval: fetchInterval,
	}
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, source := range sources {
		wg.Add(1)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] failed to fetch items from source %q: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR] failed to process items from source %q: %v", source.Name(), err)
				return
			}
		}(src.NewRSSSourceFromModel(source))
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if err := f.articles.Store(ctx, model.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}

	return nil
}
