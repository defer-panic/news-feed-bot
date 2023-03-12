package source

import (
	"context"

	"github.com/SlyMarbo/rss"
	"github.com/samber/lo"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type RSSSource struct {
	URL        string
	SourceName string
}

func (s RSSSource) Fetch(ctx context.Context) ([]model.Item, error) {
	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, err
	}

	return lo.Map(feed.Items, func(item *rss.Item, _ int) model.Item {
		return model.Item{
			Title:      item.Title,
			Link:       item.Link,
			Date:       item.Date,
			SourceName: s.SourceName,
		}
	}), nil
}

func (s RSSSource) Name() string {
	return s.SourceName
}

func (s RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	// load rss feed with context:
	var (
		feedCh = make(chan *rss.Feed)
		errCh  = make(chan error)
	)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			errCh <- err
			return
		}
		feedCh <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case feed := <-feedCh:
		return feed, nil
	}
}
