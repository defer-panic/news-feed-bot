package fetcher

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-shiori/go-readability"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type ArticleStorage interface {
	StoreArticle(ctx context.Context, article model.ClassifiedItem) error
	GetTopArticles(
		ctx context.Context,
		chatTelegramID int64,
		topics []string,
		n int64,
		timeWindow time.Duration,
	) ([]model.ClassifiedItem, error)
}

type Fetcher struct {
	mu              sync.Mutex
	itemsClassified sync.Map
	rand            *rand.Rand
	config          Config
}

type Config struct {
	Sources                      []Source
	FetchInterval                time.Duration
	ContentExtractInterval       time.Duration
	ContentExtractIntervalJitter time.Duration
	Classifier                   Classifier
	ArticleStorage               ArticleStorage
}

func New(config Config) *Fetcher {
	return &Fetcher{config: config}
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.config.FetchInterval)
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
	log.Printf("[INFO] fetching items from %d sources", len(f.config.Sources))
	f.mu.Lock()
	defer f.mu.Unlock()

	f.itemsClassified = sync.Map{}
	f.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	var wg sync.WaitGroup

	for _, source := range f.config.Sources {
		wg.Add(1)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] failed to fetch items from source %q: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, items); err != nil {
				log.Printf("[ERROR] failed to process items from source %q: %v", source.Name(), err)
				return
			}
		}(source)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, items []model.Item) error {
	for _, item := range items {
		itemContent, err := f.loadItemContent(ctx, item)
		if err != nil {
			log.Printf("[ERROR] failed to load item content: %v", err)
			continue
		}

		item.Date = item.Date.UTC()

		classifyResult, err := f.config.Classifier.Classify(itemContent)
		if err != nil {
			log.Printf("[ERROR] failed to classify item: %v", err)
			continue
		}

		if err := f.config.ArticleStorage.StoreArticle(ctx, model.ClassifiedItem{
			Item:  item,
			Slug:  classifyResult.Class.Slug,
			Score: classifyResult.Score,
		}); err != nil {
			return err
		}

		log.Printf("[INFO] %+v", classifyResult)

		jitter := time.Duration(f.rand.Int63n(int64(f.config.ContentExtractIntervalJitter)))
		time.Sleep(f.config.ContentExtractInterval + jitter)
	}

	return nil
}

//func (f *Fetcher) storeClassifyResult(ctx context.Context, item model.Item, result *model.ClassifyResult) {
//	classMembers, exist := f.itemsClassified.Load(result.Class.Slug)
//	if !exist {
//		classMembers = make([]model.ClassifiedItem, 0, 1)
//	}
//
//	typedClassMembers := classMembers.([]model.ClassifiedItem)
//	typedClassMembers = append(typedClassMembers, model.ClassifiedItem{
//		Item:  item,
//		Slug:  result.Class.Slug,
//		Score: result.Score,
//	})
//
//	f.itemsClassified.Store(result.Class.Slug, typedClassMembers)
//
//	for _, classifiedItem := range typedClassMembers {
//		if err := f.config.ArticleStorage.StoreArticle(ctx, classifiedItem); err != nil {
//			log.Printf("[ERROR] failed to store article: %v", err)
//		}
//	}
//
//}

func (f *Fetcher) itemsClassifiedToMap() map[string][]model.ClassifiedItem {
	itemsClassifiedMap := make(map[string][]model.ClassifiedItem)
	f.itemsClassified.Range(func(key, value any) bool {
		itemsClassifiedMap[key.(string)] = value.([]model.ClassifiedItem)
		return true
	})

	return itemsClassifiedMap
}

func (f *Fetcher) loadItemContent(ctx context.Context, item model.Item) (string, error) {
	var (
		textCh = make(chan string)
		errCh  = make(chan error)
	)

	go func() {
		article, err := readability.FromURL(item.Link, 30*time.Second)
		if err != nil {
			errCh <- err
			return
		}
		textCh <- article.Content
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errCh:
		return "", err
	case text := <-textCh:
		return text, nil
	}
}

type Result struct {
	ItemsClassified map[string][]model.ClassifiedItem
}

type Source interface {
	Fetch(ctx context.Context) ([]model.Item, error)
	Name() string
}

type Classifier interface {
	Classify(text string) (*model.ClassifyResult, error)
}
